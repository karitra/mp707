
Remarks
=======

1. Don't forget to add [/etc/udev/rules.d/80-mp707.rules] file with following line:

ATTR{idVendor}=="16c0", ATTR{idProduct}=="05df", MODE="0666"

It allows to access from libusb to non-root users on udev based systems.
At least it works for me on Ubuntu 12.04.

You can use more sophisticated rule set with group access, see online docs for examples [1].

Footnote:
1. udev - Linux dynamic device management [https://wiki.debian.org/udev] 