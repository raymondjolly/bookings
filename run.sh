#!/bin/bash

go build -o bookings cmd/web/*.go
./bookings -dbname=bookings -dbuser=raymondjolly -cache=false -production=false