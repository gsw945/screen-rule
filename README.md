# screen-rule

## notes

```bash
mkdir screen-rule
cd screen-rule
go mod init screen-rule
touch main.go
go mod tidy -v
go build -v -o screen-rule.exe main.go
screen-rule.exe
# update all packages
go get -u -t ./...
```

## assets
- [`assets/fonts/SourceHanSansSC-VF.ttf`](https://raw.githubusercontent.com/adobe-fonts/source-han-sans/release/Variable/TTF/SourceHanSansSC-VF.ttf)