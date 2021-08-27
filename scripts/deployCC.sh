source ./scripts/vars.sh

cd $FABRIC_SAMPLE_DIR

/bin/bash ./network.sh deployCC -ccn $CHAINCODE_NAME -ccp $CHAINCODE_PATH -ccl $CHAINCODE_LANGUAGE


if [ $? ]
then 
    echo "smartcontract deployed successfully"
else
    echo "error deploying smart contract"
fi

cd $WD