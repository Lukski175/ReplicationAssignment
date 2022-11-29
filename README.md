# ReplicationAssignment
To start the program, you must first setup the amount of backup servers you want.

You do so by starting a backup application, using the command:
"go run backup\backup.go {id}", 
where the id should start with 0, counting up by 1 for each backup you start.

The client is then started with:
"go run client\client.go {backupAmount}", 
where the backup amount is of course the amount of backups you have started. (The command argument should be without {})

You can then try out the program, by issuing commands from the client side, by following instructions as prompted. 
(and yes, the responses from the servers are not very clear)
