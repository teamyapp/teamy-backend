$paths = @('core','apps')

foreach ($path in $paths)
{
  $protoFiles = Get-ChildItem -Path $path -Filter *.proto -File -Recurse -Name
  foreach ($protoFile in $protoFiles) {
      protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative "$path\$protoFile"
  }
}

