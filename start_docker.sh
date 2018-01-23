#!/bin/bash

docker run --name POSTGRES -v /home/agustin/gocode/src/arthurgustin.fr/teddycare/sql/:/home/agustin/gocode/src/arthurgustin.fr/teddycare/sql/ -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=teddycare postgres
