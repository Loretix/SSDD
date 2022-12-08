docker rmi localhost:5001/cliente:latest
docker rmi localhost:5001/servidor:latest
kubectl delete statefulset ss
kubectl delete service ss-service
kubectl delete pod c1
kubectl delete service prueba
kind delete cluster