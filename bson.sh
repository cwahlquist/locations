#!/bin/bash
sed -i 's/json:"-"/json:"-" bson:"-"/g' $1
