# pi-salt
## PiSugar golang version
Refer: https://github.com/PiSugar/PiSugar

运行之前

    sudo hciconfig
    sudo hciconfig hci0 down  # or whatever hci device you want to use

If you have BlueZ 5.14+ (or aren't sure), stop the built-in
bluetooth server, which interferes with gatt, e.g.:

    sudo service bluetooth stop

