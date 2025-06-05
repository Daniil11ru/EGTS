import argparse
import pandas as pd
import psycopg2
import sys

LIMIT = 4294967295

def tails(imei: str):
    s = str(imei).strip()
    max_k = min(10, len(s))
    for k in range(max_k, 0, -1):
        tail = s[-k:]
        if int(tail) <= LIMIT:
            yield tail

def main():
    p = argparse.ArgumentParser()
    p.add_argument('--xlsx', required=True)
    p.add_argument('--db-host', required=True)
    p.add_argument('--db-port', type=int, default=5432)
    p.add_argument('--db-name', required=True)
    p.add_argument('--db-user', required=True)
    p.add_argument('--db-pass', required=True)
    args = p.parse_args()

    df = pd.read_excel(args.xlsx, dtype={'IMEI': str})
    if 'IMEI' not in df.columns:
        print('Колонка IMEI не найдена', file=sys.stderr)
        sys.exit(1)

    imeis = df['IMEI'].dropna().unique()
    total = len(imeis)
    conn = psycopg2.connect(
        host=args.db_host,
        port=args.db_port,
        dbname=args.db_name,
        user=args.db_user,
        password=args.db_pass
    )
    cur = conn.cursor()
    rows = []
    found_count = 0
    bar_len = 40

    for idx, imei in enumerate(imeis, start=1):
        filled_len = int(bar_len * idx / total)
        bar = '#' * filled_len + '-' * (bar_len - filled_len)
        print(f'\rProgress: |{bar}| {idx}/{total}', file=sys.stderr, end='', flush=True)
        match = None
        for t in tails(imei):
            cur.execute(
                "SELECT 1 FROM point WHERE point->>'client' = %s LIMIT 1",
                (t,)
            )
            if cur.fetchone():
                match = t
                found_count += 1
                break
        rows.append({
            'IMEI': imei,
            'OID': match if match is not None else '',
        })

    print(file=sys.stderr)
    conn.close()

    out = pd.DataFrame(rows, columns=['IMEI', 'OID'])
    print(out.to_string(index=False))
    print(f'Найдено {found_count} из {total} IMEI')

if __name__ == '__main__':
    main()
