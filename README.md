# 📂 fzd

> Golang file indexer and fuzzy file finder utiliy tool

## 🔨 Build

```
go install ./cmd/fzd
```

## 🛠 Usage

Copy `.fzd.example.yaml` and place it in as `$HOME/.fzd/.fzd.yaml` in your home directory.

Tweak the configurations for your case, may take a look at [detailed configuration list](#%E2%9A%99-configuration).

```
$ fzd
Index is not created yet
Do you want to create it now [y/N]: y
Indexed for 102 files

$ fzd test
/home/test.json
/home/Projects/zzz-test
/home/Projects/zzz_test
/home/Projects/file_tests
/home/Projects/more_file_tests

$ fzd -n 3 test
/home/test.json
/home/Projects/zzz-test
/home/Projects/zzz_test

$ fzd -num 3 test
/home/test.json
/home/Projects/zzz-test
/home/Projects/zzz_test
```

## ⚙ Configuration

> Coming soon

## 🚢 Release

```
go install github.com/mitchellh/gox@latest

gox -output ./build/{{.Dir}}_{{.OS}}_{{.Arch}} ./cmd/fzd
```

## 📜 License

Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.