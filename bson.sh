#!/bin/bash
cat $1 | \
awk 'BEGIN { FS=OFS="`" } { 
  if(NF>2) { 
    split( $2, array, " " ); 
    bson=""
    for (i in array) { 
      n=index(array[i],"json:") 
      if (n>0) {
        bson=array[i] 
        sub(/json/, "bson", bson) 
      }
    }
    if (length(bson)) {
        print $1" `"$2" "bson"`"
    }
  } else 
    { print }
  }' 
