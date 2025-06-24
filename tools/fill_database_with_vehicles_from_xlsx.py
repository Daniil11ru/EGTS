import argparse
import pandas as pd
import psycopg2
import sys

def get_oid(imei, oid_type, oid_from, oid_count):
    s = ''.join(ch for ch in str(imei) if ch.isdigit())
    if not s:
        raise ValueError('IMEI не содержит цифр')
    if oid_type == 'digits':
        if oid_count is None:
            raise ValueError('oid_count обязателен, если oid_type имеет значение digits')
        sub = s[:oid_count] if oid_from == 'start' else s[-oid_count:]
        return int(sub)
    if oid_type == 'bytes':
        if oid_count is None:
            raise ValueError('oid_count обязателен, если oid_type имеет значение bytes')
        int_val = int(s)
        byte_len = max((int_val.bit_length() + 7) // 8, 1)
        b = int_val.to_bytes(byte_len, 'big')
        slice_bytes = b[:oid_count] if oid_from == 'start' else b[-oid_count:]
        return int.from_bytes(slice_bytes, 'big')
    if oid_type == 'max_digits':
        max_val = 4294967295 # В случае, если беззнаковое
        for n in range(len(s), 0, -1):
            sub = s[:n] if oid_from == 'start' else s[-n:]
            if int(sub) <= max_val:
                return int(sub)
        raise ValueError('Невозможно вместиться в 4 байта')
    raise ValueError('Неизвестный oid_type')

def main():
    p = argparse.ArgumentParser()
    p.add_argument('--provider-id', type=int, required=True)
    p.add_argument('--db-host', required=True)
    p.add_argument('--db-port', type=int, default=5432)
    p.add_argument('--db-name', required=True)
    p.add_argument('--db-user', required=True)
    p.add_argument('--db-password', required=True)
    p.add_argument('--imei-column', required=True)
    p.add_argument('--plate-column', required=True)
    p.add_argument('--excel-path', required=True)
    p.add_argument('--oid-type', choices=['digits', 'bytes', 'max_digits'], required=True)
    p.add_argument('--oid-from', choices=['start', 'end'], required=True)
    p.add_argument('--oid-count', type=int)
    args = p.parse_args()

    df = pd.read_excel(args.excel_path)
    if args.imei_column not in df.columns or args.plate_column not in df.columns:
        sys.exit(f'Колонки не найдены в Excel-файле: {args.imei_column}, {args.plate_column}')
    rows = df[[args.imei_column, args.plate_column]].dropna()
    conn = psycopg2.connect(host=args.db_host, port=args.db_port, dbname=args.db_name, user=args.db_user, password=args.db_password)
    cur = conn.cursor()
    insert_sql = 'insert into vehicle (imei, oid, license_plate_number, provider_id, moderation_status) values (%s,%s,%s,%s,%s)'
    for _, r in rows.iterrows():
        imei = str(r[args.imei_column])
        plate = str(r[args.plate_column])
        try:
            oid_val = get_oid(imei, args.oid_type, args.oid_from, args.oid_count)
            cur.execute(insert_sql, (imei, oid_val, plate, args.provider_id, 'pending'))
        except Exception as exc:
            print(f'Не получилось обработать строку {imei} {plate}: {exc}', file=sys.stderr)
    conn.commit()
    cur.close()
    conn.close()

if __name__ == '__main__':
    main()