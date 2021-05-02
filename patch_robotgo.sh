#!/usr/bin/env bash
# Patches the robotgo C code to support the GRALT key on Linux systems.
set -euo pipefail

go get github.com/go-vgo/robotgo

rebuild=false
ROBOTGO_DIR=$(readlink -f $(go env GOPATH)/pkg/mod/github.com/go-vgo/robotgo@v*)

chmod -R u+w $ROBOTGO_DIR/key

[ -f $ROBOTGO_DIR/key/keycode.h.orig ] || cp $ROBOTGO_DIR/key/keycode.h $ROBOTGO_DIR/key/keycode.h.orig
if ! grep K_GRALT $ROBOTGO_DIR/key/keycode.h > /dev/null; then
    sed -i 's/K_RALT = XK_Alt_R,/\0\nK_GRALT = XK_ISO_Level3_Shift,/' $ROBOTGO_DIR/key/keycode.h
    # These are only needed so it compiles (since goKey.h will reference a K_GRALT):
    sed -i 's/K_RALT = kVK_RightOption,/\0\nK_GRALT = kVK_RightOption,/' $ROBOTGO_DIR/key/keycode.h
    sed -i 's/K_RALT = VK_RMENU,/\0\nK_GRALT = VK_RMENU,/' $ROBOTGO_DIR/key/keycode.h

    echo "Patched $ROBOTGO_DIR/key/keycode.h:"
    diff $ROBOTGO_DIR/key/keycode.h.orig $ROBOTGO_DIR/key/keycode.h || true
    rebuild=true
fi

[ -f $ROBOTGO_DIR/key/goKey.h.orig ] || cp $ROBOTGO_DIR/key/goKey.h $ROBOTGO_DIR/key/goKey.h.orig
if ! grep K_GRALT $ROBOTGO_DIR/key/goKey.h > /dev/null; then
    sed -i 's/K_RALT \},/\0\n{ "gralt", K_GRALT },/' $ROBOTGO_DIR/key/goKey.h
    echo "Patched $ROBOTGO_DIR/key/goKey.h:"
    diff $ROBOTGO_DIR/key/goKey.h.orig $ROBOTGO_DIR/key/goKey.h || true
    rebuild=true
fi

if [ $rebuild == true ]; then
    echo -n "Patched robotgo. Rebuilding..."
    go clean -cache
    go build github.com/go-vgo/robotgo
    echo "Done."
fi