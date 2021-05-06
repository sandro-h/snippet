# snippet

[![CI](https://github.com/sandro-h/snippet/actions/workflows/ci.yml/badge.svg)](https://github.com/sandro-h/snippet/actions/workflows/ci.yml)

Tiny cross-platform widget to insert snippets into the terminal, editor, etc.

![Snippet](screenshot.png)

## Usage

1. `Alt + q` to show widget
2. Start typing in search box to find snippet (fuzzy search)
3. Use `up` and `down` arrows to navigate search results
4. Press `enter` to choose snippet. Widget disappears and snippet is typed in active window.
5. Press `escape` to cancel search and hide widget again.
6. Press `Alt + F4` while widget is active to close it for good.

Snippets are stored in `snippet.yml` file, see [snippet_sample.yml](snippet_sample.yml).

Optional configuration is stored in `config.yml` file, see [config_sample.yml](config_sample.yml).

### Snippet arguments

Snippets can use arguments that you have to fill out before it is typed.
In this case:

1. When you press `enter` to select a snippet with arguments,
a new window pops up to fill out the arguments.
2. Use `up` and `down` arrows to jump between argument inputs
3. Press `enter` to confirm the arguments and type the snippet. Any empty arguments lead to empty replacements in the snippet.
4. Press `escape` to cancel and return to the main snippet window.

See the [snippet_sample.yml](snippet_sample.yml) for configuring a snippet with arguments.

## Development

```shell
make build-linux
make build-windows
make test
make lint
```

See `Makefile`'s `install-sys-packages` for required system packages for compilation.
