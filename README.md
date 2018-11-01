Smart Wallet
=============

Smart wallet is a peer-to-peer wallet where it does not use a central server/database. 

In order to send money to another wallet, both the wallets first need to be synced to each other. Then the sender creates an encrypted token which contains the:
- sender's id
- receiver's id
- amount
- counter

The receiver on the other hand confirms that the counter received and the counter saved is the same, to prevent the same token from being re-used multiple times.


To run the program 
=============
1. Make sure docker-compose is installed.
2. Run `docker-compose up` from the directory to initialise the program.
3. Navigate to `localhost:8080` to view the program.
+ Note: If port `8080` is not available you can change it in the docker-compose file.
+ Since it is using redis, in order to purge the data and start from scratch you should run `docker-compose down`and then re-run `docker-compose up`.
