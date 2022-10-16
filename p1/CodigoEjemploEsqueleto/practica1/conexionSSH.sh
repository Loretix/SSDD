#script al que se le pasa por $1 fichero con ips
#!/bin/bash

if [ $# -eq 1 ]; then
	while read ipPuerto ; do
		IP=$( echo $ipPuerto | cut -d ":" -f1 )
		puerto=$( echo $ipPuerto | cut -d ":" -f2 )
		echo $IP y $puerto
		#ejecuta en una maquina remota un fich que tienes en tu maquina
		#ssh a796598@$IP 'bash -s' < worker $ipPuerto
		#scp -i ~/.shh/id_as_ed25519 worker.go a796598@$ipPuerto:/SSDD/practica1 
		#ssh -n a796598@$IP  'export PATH=$PATH:/usr/local/go/bin;go run $HOME/SSDD/practica1/worker.go $ipPuerto' 
		ssh -n a796598@$IP  "cd SSDD/practica1;/usr/local/go/bin/go run worker.go $IP $puerto"
		#ssh -n -i ~/.shh/id_as_ed25519 a796598@$ipPuerto ''
     done < $1
else
    echo "Numero incorrecto de parametros"
fi	  


