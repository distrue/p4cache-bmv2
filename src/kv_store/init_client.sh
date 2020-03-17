usage="${0} [<server-init-flags>]"

server_flags=$1

mx client1 python3 ./client/exec_queries.py $server_flags
