$scriptpath = $MyInvocation.MyCommand.Path
$dir = Split-Path $scriptpath
Push-Location $dir/..

$key = Read-Host -Prompt "Please provide your API Key"
$rootUrl = Read-Host -Prompt "Please provide the root URL(s)"

$env:GRAFANA_API_KEY = $key
npx @grafana/toolkit plugin:sign --rootUrls $rootUrl

Pop-Location