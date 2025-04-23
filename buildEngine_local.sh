#!/bin/bash
rm -rf .engines
mkdir .engines
cd .engines
git clone https://github.com/eclipse-xfsc/crypto-provider-local-plugin.git .local
git clone https://github.com/eclipse-xfsc/crypto-provider-hashicorp-vault-plugin.git .vault

cd .local
rm localProvider_test.go
go build -buildmode=plugin -gcflags="all=-N -l"
cd .. 

cd .vault
rm vaultProvider_test.go
go build -buildmode=plugin -gcflags="all=-N -l"
cd ..
