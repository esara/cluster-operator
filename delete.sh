#!/bin/bash

kubectl delete -f nfs.yaml &
kubectl delete nfsserver --all &
kubectl delete statefulset --all &
kubectl delete pvc --all &
kubectl delete pv --all &
storageos v rm --all &

