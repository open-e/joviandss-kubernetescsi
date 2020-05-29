#!/usr/bin/bash

rm ./bins/jdss-csi-plugin
rm ./bins/joviandss-kubernetes-csi-latest

cp ./build/src/_output/* ./bins/
