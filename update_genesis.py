#!/usr/bin/env python3

import json

# Читаем genesis файл
with open('/Users/gp/.nuahd/config/genesis.json', 'r') as f:
    genesis = json.load(f)

# Добавляем секцию freeaccount с адресом testuser
genesis['app_state']['freeaccount'] = {
    "free_accounts": [
        "nuah1mlla4jl6hk04urr55vgkn83wtjpmunlvg3sccj"
    ]
}

# Записываем обновленный файл
with open('/Users/gp/.nuahd/config/genesis.json', 'w') as f:
    json.dump(genesis, f, indent=2)

print("Genesis file updated successfully!")