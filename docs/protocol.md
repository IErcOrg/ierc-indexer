
## IERC protocol json

```json
{
  "ierc-pow:deploy": {
    "p": "ierc-pow",
    "op": "deploy",
    "tick": "ethpi",
    "max": "21000000000",
    "dec": "2",
    "tokenomics": {
      "18992800": "1000",
      "24248800": "500",
      "29504800": "250"
    },
    "rule": {
      "pow": 50,
      "min_workc": "0x0000",
      "difficulty_ratio": 10,
      "pos": 50,
      "pool": "0x0000000000000000000000000000000000000000",
      "max_reward_block": 5
    }
  },
  "ierc-pow:mint": {
    "p": "ierc-pow",
    "op": "mint",
    "tick": "ethpi",
    "use_point": "500",
    "block": "1233",
    "nonce": "1111"
  },
  "ierc-pow:modify": {
    "p": "ierc-pow",
    "op": "modify",
    "tick": "ethpi",
    "max": "100000000"
  },
  "ierc-pow:airdrop_claim": {
    "p": "ierc-pow",
    "op": "airdrop_claim",
    "tick": "ethpi",
    "claim": "3600000" 
  },
  "ierc-20:deploy": {
    "p": "ierc-20",
    "op": "deploy",
    "tick": "ethi",
    "max": "21000000",
    "lim": "1000",
    "wlim": "10000",
    "dec": "8",
    "nonce": "10"
  },
  "ierc-20:mint": {
    "p": "ierc-20",
    "op": "mint",
    "tick": "ethi",
    "amt": "1000",
    "nonce": "11"
  },
  "ierc-20|ierc-pow:transfer": {
    "p": "ierc-20",
    "op": "transfer",
    "tick": "ethi",
    "nonce": "45",
    "to": [
      {
        "recv": "0x7BBAF8B409145Ea9454Af3D76c6912b9Fb99b2A9",
        "amt": "10000"
      }
    ]
  },
  "ierc-20|ierc-pow:freeze_sell": {
    "p": "ierc-20",
    "op": "freeze_sell",
    "freeze": [
      {
        "amt": "300",
        "gasPrice": "28785280",
        "nonce": "1704977704728",
        "platform": "0x1878d3363a02f1b5e13ce15287c5c29515000656",
        "seller": "0x4444777786851a1b941a86694f5f9a11da070f3f",
        "ssign": "0x9bae3b45c3e028d92f445047819c3062d7a4669c4a2551175c199ca77c7c95f61862419fa9e81f71464ac5d137365546d80b1b69e32b2a63d090cbabda8269c21c",
        // "buyer": "0x2222777786851a1b941a86694f5f9a11da070f3f",
        // "bsign":"0x9bae3b45c3e028d92f445047819c3062d7a4669c4a2551175c199ca77c7c95f61862419fa9e81f71464ac5d137365546d80b1b69e32b2a63d090cbabda8269c21c",
        "tick": "BTCI",
        "payment": "ethi",
        "value": "0.0054"
      }
    ]
  },
  "ierc-20|ierc-pow:unfreeze_sell": {
    "p": "ierc-20",
    "op": "unfreeze_sell",
    "unfreeze": [
      {
        "txHash": "0x649ad8221d03891ecd7426fb26fc239124e7b7dd042c57e1f8cc43fc99b379f3",
        "sign": "0x3d6a75e49d15e4210940db48e666c1e0071eca92250edfb2318e0b83081667613083b99bc779e7546dd80ecab2f399a23a7488eacb46514049ba10a4c116cf921b",
        "msg": "order record status not is list or pending (freeze sell)"
      }
    ]
  },
  "ierc-20|ierc-pow:proxy_transfer": {
    "p": "ierc-20",
    "op": "proxy_transfer",
    "proxy": [
      {
        "tick": "ethi",
        "nonce": "20",
        "from": "0x22222222222222222222222222222222222222222222",
        "to": "0x22222222222222222222222222222222222222222222",
        "amt": "333",
        "value": "0.001",
        "sign": "0x000"
      }
    ]
  },
  "ierc-20:stake_config": {
    "p": "ierc-20",
    "op": "stake_config",
    "name": "abcdef",
    "pool": "0x9D2576874311932Ef54e9f9ff4aeff7Cd7Ab1419",
    "id": "1",
    "owner": "0x9D2576874311932Ef54e9f9ff4aeff7Cd7Ab1419",
    "details": [
      {
        "tick": "ethi",
        "ratio": "0.01",
        "max_amt": "100000"
      }
    ],
    "stop_block": "100000"
  },
  "ierc-20:stake": {
    "p": "ierc-20",
    "op": "stake",
    "pool": "0x0000000000000000000000000000000000000000",
    "id": "1",
    "details": [
      {
        "tick": "ethi",
        "amt": "10000"
      }
    ]
  },
  "ierc-20:unstake": {
    "p": "ierc-20",
    "op": "unstake",
    "pool": "0x0000000000000000000000000000000000000000",
    "id": "1",
    "details": [
      {
        "staker": "0x00001",
        "tick": "ethi",
        "amt": "10000"
      }
    ]
  },
  "ierc-20:proxy_unstake": {
    "p": "ierc-20",
    "op": "proxy_unstake",
    "pool": "0x0000000000000000000000000000000000000000",
    "id": "1",
    "details": [
      {
        "staker": "0x00001",
        "tick": "ethi",
        "amt": "10000"
      }
    ]
  }
}

```
