import argparse
import pandas as pd
import psycopg2
import sys

LIMIT = 4294967295

def tails(imei):
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

    imeis = df['IMEI'].dropna().str.strip().unique()
    conn = psycopg2.connect(
        host=args.db_host,
        port=args.db_port,
        dbname=args.db_name,
        user=args.db_user,
        password=args.db_pass
    )
    cur = conn.cursor()
    rows = []

    for imei in imeis:
        match = ''
        for tail in tails(imei):
            cur.execute(
                "SELECT 1 FROM point WHERE (point->>'client')::bigint = %s LIMIT 1",
                (int(tail),)
            )
            if cur.fetchone():
                match = tail
                break
        rows.append({
            'IMEI': imei,
            'client_tail': match,
        })

    conn.close()

    out = pd.DataFrame(rows, columns=['IMEI', 'client_tail'])
    print(out.to_string(index=False))

if __name__ == '__main__':
    main()
