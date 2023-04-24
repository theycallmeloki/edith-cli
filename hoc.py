import json
import os

chain_types= ["arb", "eth"]

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
            print(f"./edith arb --wallet {i} --chain {j} | jq -s . | tee scratch/generated/{i}_{j}_{k}.json")
            print(f"echo Progress: {i} {str(holders[k].index(i))}/{str(len(holders[k]))} {j} {k} attempting to be generated")
