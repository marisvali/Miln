$Env:GOOS = 'js'
$Env:GOARCH = 'wasm'
go build -tags online,fixedlevels,int_overflow_checks_disabled,world_debug_info_disabled -o miln999.wasm github.com/marisvali/miln
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

$client = New-Object System.Net.WebClient
$client.Credentials = New-Object System.Net.NetworkCredential($Env:MILN_FTP_USER, $Env:MILN_FTP_PASSWORD)
$client.UploadFile("ftp://ftp.playful-patterns.com/public_html/miln999.wasm", "miln999.wasm")