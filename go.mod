module github.com/sandro-h/snippet

go 1.16

require (
	fyne.io/fyne/v2 v2.0.3
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-vgo/robotgo v0.93.1
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/robotn/gohook v0.30.6
	github.com/sahilm/fuzzy v0.1.0
	github.com/sosedoff/ansible-vault-go v0.0.0-20201201002713-782dc5c40224
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/go-vgo/robotgo => github.com/sandro-h/robotgo v0.99.0-linuxfix3
