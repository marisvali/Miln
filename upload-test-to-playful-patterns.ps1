$Env:GOOS = 'js'
$Env:GOARCH = 'wasm'
go build -o miln-test.wasm github.com/marisvali/miln/research/setmouse
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

$client = New-Object System.Net.WebClient
$client.Credentials = New-Object System.Net.NetworkCredential($Env:MILN_FTP_USER, $Env:MILN_FTP_PASSWORD)
$client.UploadFile("ftp://ftp.playful-patterns.com/public_html/miln-test.wasm", "miln-test.wasm")