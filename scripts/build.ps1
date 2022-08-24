$scriptpath = $MyInvocation.MyCommand.Path
$dir = Split-Path $scriptpath
Push-Location $dir/..

yarn install --pure-lockfile
yarn build
go run mage.go -v

Pop-Location