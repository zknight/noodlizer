cp noodlizer /usr/local/bin/
if [ ! -d "/usr/local/etc/noodlizer" ]; then
    echo "noodlizer data dir not exists. Creationing..."
    mkdir /usr/local/etc/noodlizer
    chown /usr/local/etc/noodlizer noodlizer
fi
cp -r template /usr/local/etc/noodlizer/template/
cp -r static /usr/local/etc/noodlizer/static/
if [ ! -a "/usr/local/etc/noodlizer/tracks.db" ]; then
    echo "noodlizer db don't be. copy it..."
    cp tracks.db /usr/local/etc/noodlizer/tracks.db
fi