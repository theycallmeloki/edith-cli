python3 hoc.py > run.sh

export ARBI_API_KEY=$(cat scratch/apikeys/arbiApiKey.txt)
export ETHER_API_KEY=$(cat scratch/apikeys/etherApiKey.txt)

echo '----------------------------------------'
echo 'dropping previous configuration file' 
rm -rf ~/.config/edith
rm -rf scratch/generated/*
echo '----------------------------------------'

echo '----------------------------------------'
echo 'creating new test build of edithctl'
go build -o edith -ldflags="-X 'github.com/theycallmeloki/edith-cli/cmd/edithctl.version=0.0.1'" main.go
echo '----------------------------------------'

echo '----------------------------------------'
echo 'checking for edith in configuration directory'
ls -1 ~/.config | grep edith
echo '----------------------------------------'

echo '----------------------------------------'
echo 'CMD: edith --help'
# ./edith --help
echo '----------------------------------------'

echo '----------------------------------------'
echo 'CMD: edith configure arbiApiKey'
echo "$ARBI_API_KEY" | ./edith configure --arbiApiKey -
echo '----------------------------------------'

echo '----------------------------------------'
echo 'CMD: edith configure etherApiKey'
echo "$ETHER_API_KEY" | ./edith configure --etherApiKey -
echo '----------------------------------------'

echo '----------------------------------------'
echo 'display current configuration file'
cat ~/.config/edith/edith.json
echo '----------------------------------------'

echo '----------------------------------------'
echo 'CMD: edith arb --help'
# ./edith arb --help
echo '----------------------------------------' 

echo '----------------------------------------'
echo 'CMD: edith arb --wallet wallet_address'
./edith arb --wallet wallet_address
echo '----------------------------------------' 