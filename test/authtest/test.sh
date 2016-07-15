#!/bin/sh

cd ./organizationtest && go test -v && cd ..
cd ./repositorytest && go test -v && cd ..
cd ./teamtest && go test -v && cd ..
cd ./usertest && go test -v && cd ..
cd ./ldaptest && go test -v && cd ..
go test -v

