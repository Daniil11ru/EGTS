#!/usr/bin/env python3

import os
import re
import argparse

def process_file(path, pattern):
    with open(path, 'r', encoding='utf-8') as f:
        text = f.read()
    fixes = 0
    def repl(m):
        nonlocal fixes
        s = m.group(1)
        if s and s[0].isupper():
            fixes += 1
            return m.group(0).replace(s, s[0].lower() + s[1:], 1)
        return m.group(0)
    new = pattern.sub(repl, text)
    if new != text:
        with open(path, 'w', encoding='utf-8') as f:
            f.write(new)
    return fixes

def main():
    parser = argparse.ArgumentParser(description='Fix first letter in fmt.Errorf messages')
    parser.add_argument('root', nargs='?', default='.', help='directory to scan')
    args = parser.parse_args()

    pattern = re.compile(r'fmt\.Errorf\(\s*"([^"]*)"', re.MULTILINE)
    total_fixes = 0

    for dirpath, _, files in os.walk(args.root):
        for name in files:
            if name.endswith('.go'):
                total_fixes += process_file(os.path.join(dirpath, name), pattern)

    print(total_fixes)

if __name__ == '__main__':
    main()
