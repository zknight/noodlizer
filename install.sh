go build
if [ $? -eq 0 ]; then
    echo "Build success. Installing at /usr/local/bin and /usr/local/etc/noodlizer"
    cp noodlizer /usr/local/bin/
    if [ ! -d "/usr/local/etc/noodlizer" ]; then
        mkdir /usr/local/etc/noodlizer
    fi
fi