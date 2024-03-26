# ierc-indexer

The IERC 20 has introduced a comprehensive set of efficient and gas-saving indexes built on Ethereum. By utilizing hexadecimal representation, event information is permanently stored in the EVM input, ensuring tamper-proof data integrity.

When no value is sent, events are directed to the black hole address, facilitating faster information retrieval for indexing purposes. This approach optimizes the protocol's performance and enhances the efficiency of data retrieval and analysis.

This indexer is able to parse Deploy\Mint\Transfer\Proxy transfer\Freeze sell and other operations related to the IERC protocol on the chain.

## Environment Requirements

To run `ierc-indexer`, ensure your system meets the following requirements:

- **Golang**: Recommended version 1.21
- **MySQL**: Recommended version 8.0
- **Git**

## Deployment Instructions

### Clone the Source Code

Start by cloning the repository to your local machine:

```bash
git clone https://github.com/IErcOrg/ierc-indexer.git
```

### Install MySQL Server

1. **Install MySQL Server**:
   Update your package index and install the MySQL server by running:

   ```bash
   apt update
   apt install mysql-server
   ```

2. **Create the Database**:
   Log into your MySQL instance and create a new database named `indexer`.

3. **Edit Configuration File**:
   Update the configuration file with your database details:

   ```yaml
   data:
     database:
       driver: mysql
       log_level: 2
       source: "$USER:$PASSWORD@($HOST:$PORT)/indexer?charset=utf8mb4&parseTime=True&loc=Local"
   ```

   Replace:

   - `$USER` with your database username
   - `$PASSWORD` with your database password
   - `$HOST` with your database host address
   - `$PORT` with your database service port

### Install Golang

1. Download and install the recommended version of Golang:

```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
```

2. Add Golang to your PATH by editing your `~/.bashrc` file:

```bash
vim ~/.bashrc
```

3. Add the following line to the end of the file:

```bash
export PATH=$PATH:/usr/local/go/bin
```

4. Apply the changes:

```bash
source ~/.bashrc
```

### Compile the Program

Navigate to the cloned directory, tidy up dependencies, and compile the program:

```bash
cd ierc-indexer
go mod tidy
mkdir -p build/
go build -o ./build/indexer ./cmd/indexer/
```

### Running Index Service

Run the index service by specifying the configuration file:

```bash
./build/indexer -c configs/config.yaml
```

- `indexer`: This is the executable binary program.
- `config.yaml`: This is the configuration file used by the indexer.

## Quick Start

The indexing service primarily functions to automatically fetch blocks, clean data, and save it to a local database. It provides the following 2 API query interfaces:

1. Query the current system's processing block status, returning the latest processed block number.

   - Endpoint: http://127.0.0.1:12300/api/v1/index/status
   - Request Method: GET
   - Request Parameters: None
   - Response Example:
     ```json
     {
       "syncBlock": 19373473
     }
     ```

2. Query block events, returning event details for multiple block numbers.

   - Endpoint: http://127.0.0.1:12300/api/v1/index/events
   - Request Method: GET
   - Request Parameters:
     ```json
     {
       "startBlock": 19373473,
       "size": 10
     }
     ```
   - Response Example:
     ```json
     {
       "eventByBlocks": [
         {
           "blockNumber": "19373498",
           "prevBlockNumber": "19373462",
           "events": [
             {
               "blockNumber": "19373498",
               "txHash": "0x9f83cfcad09d8d3240e3d47f7984d203cab6db8e00f8220f6ca136937f778fe9",
               "posInIercTxs": 0,
               "from": "0xbb0a82ea45a426ed95960743b3b849947ebf7286",
               "to": "0x33302dbff493ed81ba2e7e35e2e8e833db023333",
               "value": "1430475000000000000",
               "eventAt": "1709696315000",
               "errCode": 0,
               "errReason": "",
               "tickTransferred": {
                 "protocol": "ierc-20",
                 "operate": 4,
                 "tick": "ethi",
                 "from": "0x3590f55eecc2eb731e6888acf5e5c87f22be85be",
                 "to": "0x3590f55eecc2eb731e6888acf5e5c87f22be85be",
                 "amount": "7675",
                 "ethValue": "1.4",
                 "gasPrice": "82477237483",
                 "signerNonce": "1709696165605",
                 "sign": "0x128033552cbc50a439afe7ec8a79ea797aabaa1ec88139a0b929064b8652545f6001635831444c8de483b304e43a9a94db92a14ca9425b688166b37f7d9c301c1b"
               }
             }
           ]
         }
       ]
     }
     ```

## Contribution

Contributions to `ierc-indexer` are welcome. Please ensure to follow the code of conduct and submit pull requests for any new features or bug fixes.

## License

`ierc-indexer` is licensed under the MIT License. See the LICENSE file for more details.
