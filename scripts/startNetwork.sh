#!/bin/bash

source ./scripts/vars.sh

function fabricSampleDirExists() {
    if [ -d $FABRIC_SAMPLE_DIR ]
    then 
        return 0
    else 
        return 1
    fi
}


if fabricSampleDirExists
then 
    echo "fabric-sample directory located"
else    
    echo "fabric-sample directory could not be located: $FABRIC_SAMPLE_DIR"
fi

cd $FABRIC_SAMPLE_DIR
echo $PWD

/bin/bash ./network.sh up createChannel -ca -s couchdb #restart the network

if [ $? ]
then 
    echo "network started successfully"
else
    echo "the network was unable to be started"
fi

cd $WD