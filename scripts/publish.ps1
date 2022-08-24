$scriptpath = $MyInvocation.MyCommand.Path
$dir = Split-Path $scriptpath
Push-Location $dir/..

Copy-Item -r ./dist hathoute-disruptive-datasource

$version = Read-Host -Prompt "Please provide the release version"

# Using wsl because grafana doesnt recognize archives made with Compress-Archive
wsl zip -r hathoute-disruptive-datasource-$version.zip hathoute-disruptive-datasource

Remove-Item -r hathoute-disruptive-datasource

Pop-Location