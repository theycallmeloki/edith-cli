import json
import os
from time import sleep

chain_types= ["arb", "eth", "canto"]
# chain_types = ["canto"]
# sleep(1)

nft_wallets = []
for i in os.listdir('scratch/wallets'):
    # print(i)
    if i.endswith('.txt'):
        nft_wallets.append(i.rstrip('.txt'))

# print(nft_wallets)

holders = {}
for i in nft_wallets:
    holders[f"{i}"] = []
    for j in open(f"scratch/wallets/{i}.txt"):
        holders[i].append(j.strip())

# print(json.dumps(holders, indent=4))

for k in holders.keys():
    for j in chain_types:
        for i in holders[k]:
            file_path = f"scratch/generated/{i}_{j}_{k}.json"
            print(f"if [ ! -e {file_path} ]; then")
            print(f"  edith arb --wallet {i} --chain {j} | jq -s . | tee {file_path}")
            print(f"  sleep 1.5")
            print(f"  echo Progress: {i} {str((holders[k].index(i))+1)}/{str(len(holders[k]))} {j} {k} attempting to be generated")
            print(f"else")
            print(f"  echo Skipping: {i} {str((holders[k].index(i))+1)}/{str(len(holders[k]))} {j} {k} file already exists")
            print(f"fi")
