package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/bettercap/gatt/linux/cmd"
	"log"
	"time"

	"github.com/bettercap/gatt"
	"github.com/bettercap/gatt/examples/service"
)

const (
	SERVICE_ID            = "FD2B4448AA0F4A15A62FEB0BE77A0000"
	SERVICE_NAME          = "FD2B4448AA0F4A15A62FEB0BE77A0001"
	DEVICE_MODEL          = "FD2B4448AA0F4A15A62FEB0BE77A0002"
	WIFI_NAME             = "FD2B4448AA0F4A15A62FEB0BE77A0003"
	IP_ADDRESS            = "FD2B4448AA0F4A15A62FEB0BE77A0004"
	INPUT                 = "FD2B4448AA0F4A15A62FEB0BE77A0005"
	NOTIFY_MESSAGE        = "FD2B4448AA0F4A15A62FEB0BE77A0006"
	INPUT_SEP             = "FD2B4448AA0F4A15A62FEB0BE77A0007"
	CUSTOM_COMMAND_INPUT  = "FD2B4448AA0F4A15A62FEB0BE77A0008"
	CUSTOM_COMMAND_NOTIFY = "FD2B4448AA0F4A15A62FEB0BE77A0009"
	CUSTOM_INFO_LABEL     = "FD2BCCCA"
	CUSTOM_INFO_COUNT     = "00000000000000000000FD2BCCAA0000"
	CUSTOM_INFO           = "FD2BCCCB"
	CUSTOM_COMMAND_LABEL  = "FD2BCCCC"
	CUSTOM_COMMAND_COUNT  = "00000000000000000000FD2BCCAC0000"
)

func newServices() *gatt.Service {
	service := gatt.NewService(gatt.MustParseUUID(SERVICE_ID))
	//	SERVICE_NAME
	c := service.AddCharacteristic(gatt.MustParseUUID(SERVICE_NAME))
	c.HandleReadFunc(func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
		_, _ = fmt.Fprint(rsp, "PiSugar BLE Wifi Config")
	})
	c.AddDescriptor(gatt.UUID16(2001)).SetStringValue("PiSugar BLE Wifi Config")

	//	Device Model
	c = service.AddCharacteristic(gatt.MustParseUUID(DEVICE_MODEL))
	c.HandleReadFunc(func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
		_, _ = fmt.Fprint(rsp, "Fake PI VER 1.1")
	})
	c.AddDescriptor(gatt.UUID16(2002)).SetStringValue("Raspberry Hardware Model")

	//	Wifi Name
	c = service.AddCharacteristic(gatt.MustParseUUID(WIFI_NAME))
	c.HandleNotifyFunc(func(r gatt.Request, n gatt.Notifier) {
		for !n.Done() {
			_, _ = n.Write([]byte("Fake Wifi Name"))
			time.Sleep(5 * time.Second)
		}
	})

	//	IP Address
	c = service.AddCharacteristic(gatt.MustParseUUID(IP_ADDRESS))
	c.HandleNotifyFunc(func(r gatt.Request, n gatt.Notifier) {
		for !n.Done() {
			_, _ = n.Write([]byte("1.1.1.1"))
			time.Sleep(5 * time.Second)
		}
	})

	//	Input
	c = service.AddCharacteristic(gatt.MustParseUUID(INPUT))
	c.HandleWriteFunc(func(r gatt.Request, data []byte) (status byte) {
		fmt.Println("input: " + string(data))
		return gatt.StatusSuccess
	})

	//	Input Sep
	c = service.AddCharacteristic(gatt.MustParseUUID(INPUT_SEP))
	c.HandleWriteFunc(func(r gatt.Request, data []byte) (status byte) {
		fmt.Println("input sep: " + string(data))
		return gatt.StatusSuccess
	})

	//	Input
	c = service.AddCharacteristic(gatt.MustParseUUID(CUSTOM_COMMAND_INPUT))
	c.HandleWriteFunc(func(r gatt.Request, data []byte) (status byte) {
		fmt.Println("custom command input: " + string(data))
		return gatt.StatusSuccess
	})

	//	COMMAND notify
	c = service.AddCharacteristic(gatt.MustParseUUID(CUSTOM_COMMAND_NOTIFY))
	c.HandleNotifyFunc(func(r gatt.Request, n gatt.Notifier) {
		for !n.Done() {
			_, _ = n.Write([]byte(time.Now().String()))
			time.Sleep(10 * time.Second)
		}
	})

	//	Notify Message
	c = service.AddCharacteristic(gatt.MustParseUUID(NOTIFY_MESSAGE))
	c.HandleNotifyFunc(func(r gatt.Request, n gatt.Notifier) {
		for !n.Done() {
			_, _ = n.Write([]byte(time.Now().String()))
			time.Sleep(20 * time.Second)
		}
	})

	//	Info Count
	c = service.AddCharacteristic(gatt.MustParseUUID(CUSTOM_INFO_COUNT))
	c.HandleReadFunc(func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
		_, _ = fmt.Fprint(rsp, "0")
	})
	c.AddDescriptor(gatt.MustParseUUID(CUSTOM_INFO_COUNT)).SetStringValue("Custom Info Count")

	//	Command Count
	c = service.AddCharacteristic(gatt.MustParseUUID(CUSTOM_COMMAND_COUNT))
	c.HandleReadFunc(func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
		_, _ = fmt.Fprint(rsp, "0")
	})
	c.AddDescriptor(gatt.MustParseUUID(CUSTOM_COMMAND_COUNT)).SetStringValue("Custom Info Count")

	return service
}

// server_lnx implements a GATT server.
// It uses some linux specific options for more finer control over the device.

var (
	mc    = flag.Int("mc", 1, "Maximum concurrent connections")
	id    = flag.Duration("id", 0, "ibeacon duration")
	ii    = flag.Duration("ii", 5*time.Second, "ibeacon interval")
	name  = flag.String("name", "Gopher", "Device Name")
	chmap = flag.Int("chmap", 0x7, "Advertising channel map")
	dev   = flag.Int("dev", -1, "HCI device ID")
	chk   = flag.Bool("chk", true, "Check device LE support")
)

// cmdReadBDAddr implements cmd.CmdParam for demostrating LnxSendHCIRawCommand()
type cmdReadBDAddr struct{}

func (c cmdReadBDAddr) Marshal(b []byte) {}
func (c cmdReadBDAddr) Opcode() int      { return 0x1009 }
func (c cmdReadBDAddr) Len() int         { return 0 }

// Get bdaddr with LnxSendHCIRawCommand() for demo purpose
func bdaddr(d gatt.Device) {
	rsp := bytes.NewBuffer(nil)
	if err := d.Option(gatt.LnxSendHCIRawCommand(&cmdReadBDAddr{}, rsp)); err != nil {
		fmt.Printf("Failed to send HCI raw command, err: %s", err)
	}
	b := rsp.Bytes()
	if b[0] != 0 {
		fmt.Printf("Failed to get bdaddr with HCI Raw command, status: %d", b[0])
	}
	log.Printf("BD Addr: %02X:%02X:%02X:%02X:%02X:%02X", b[6], b[5], b[4], b[3], b[2], b[1])
}

func main() {
	flag.Parse()
	d, err := gatt.NewDevice(
		gatt.LnxMaxConnections(*mc),
		gatt.LnxDeviceID(*dev, *chk),
		gatt.LnxSetAdvertisingParameters(&cmd.LESetAdvertisingParameters{
			AdvertisingIntervalMin: 0x00f4,
			AdvertisingIntervalMax: 0x00f4,
			AdvertisingChannelMap:  0x07,
		}),
	)

	if err != nil {
		log.Printf("Failed to open device, err: %s", err)
		return
	}

	// Register optional handlers.
	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { log.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { log.Println("Disconnect: ", c.ID()) }),
		gatt.PeripheralConnected(func(peripheral gatt.Peripheral, err error) {
			log.Println("p connect: ", peripheral.ID())
		}),
	)

	d.StopAdvertising()
	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			// Get bdaddr with LnxSendHCIRawCommand()
			bdaddr(d)

			// Setup GAP and GATT services.
			d.AddService(service.NewGapService(*name))
			d.AddService(service.NewGattService())

			services := newServices()
			d.AddService(services)
			uuids := []gatt.UUID{services.UUID()}

			// If id is zero, advertise name and services statically.
			if *id == time.Duration(0) {
				d.AdvertiseNameAndServices(*name, uuids)
				break
			}

			// If id is non-zero, advertise name and services and iBeacon alternately.
			go func() {
				for {
					// Advertise as a RedBear Labs iBeacon.
					d.AdvertiseIBeacon(gatt.MustParseUUID("5AFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), 0, 0, -59)
					time.Sleep(*id)

					// Advertise name and services.
					d.AdvertiseNameAndServices(*name, uuids)
					time.Sleep(*ii)
				}
			}()

		default:
		}
	}

	d.Init(onStateChanged)
	select {}
}
