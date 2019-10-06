pushd ui
  npm install
  npm run-script build
popd

statik -src=ui/build

go build -o pbdemo cmd/pbdemo/main.go
