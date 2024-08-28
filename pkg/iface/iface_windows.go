package iface

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	_ "unsafe"

	"github.com/lucheng0127/vtun/pkg/utils"
	"github.com/vishvananda/netlink"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wintun"
)

type IFace interface {
	Close() error
	Name() string
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

const (
	EventUp = 1 << iota
	EventDown
	EventMTUUpdate

	rateMeasurementGranularity = uint64((time.Second / 2) / time.Nanosecond)
	spinloopRateThreshold      = 800000000 / 8                                   // 800mbps
	spinloopDuration           = uint64(time.Millisecond / 80 / time.Nanosecond) // ~1gbit/s
)

type Event int

type rateJuggler struct {
	current       atomic.Uint64
	nextByteCount atomic.Uint64
	nextStartTime atomic.Int64
	changing      atomic.Bool
}

type NativeTun struct {
	wt        *wintun.Adapter
	name      string
	handle    windows.Handle
	rate      rateJuggler
	session   wintun.Session
	readWait  windows.Handle
	events    chan Event
	running   sync.WaitGroup
	closeOnce sync.Once
	close     atomic.Bool
	forcedMTU int
	outSizes  []int
}

var (
	WintunTunnelType          = "WireGuard"
	WintunStaticRequestedGUID *windows.GUID
)

func (tun *NativeTun) Close() error {
	var err error
	tun.closeOnce.Do(func() {
		tun.close.Store(true)
		windows.SetEvent(tun.readWait)
		tun.running.Wait()
		tun.session.End()
		if tun.wt != nil {
			tun.wt.Close()
		}
		close(tun.events)
	})
	return err
}

func (tun *NativeTun) MTU() (int, error) {
	return tun.forcedMTU, nil
}

func (tun *NativeTun) Name() string {
	return tun.name
}

//go:linkname nanotime runtime.nanotime
func nanotime() int64

//go:linkname procyield runtime.procyield
func procyield(cycles uint32)

func (rate *rateJuggler) update(packetLen uint64) {
	now := nanotime()
	total := rate.nextByteCount.Add(packetLen)
	period := uint64(now - rate.nextStartTime.Load())
	if period >= rateMeasurementGranularity {
		if !rate.changing.CompareAndSwap(false, true) {
			return
		}
		rate.nextStartTime.Store(now)
		rate.current.Store(total * uint64(time.Second/time.Nanosecond) / period)
		rate.nextByteCount.Store(0)
		rate.changing.Store(false)
	}
}

func (tun *NativeTun) Read(buf []byte) (int, error) {
	tun.running.Add(1)
	defer tun.running.Done()
retry:
	if tun.close.Load() {
		return 0, os.ErrClosed
	}
	start := nanotime()
	shouldSpin := tun.rate.current.Load() >= spinloopRateThreshold && uint64(start-tun.rate.nextStartTime.Load()) <= rateMeasurementGranularity*2
	for {
		if tun.close.Load() {
			return 0, os.ErrClosed
		}
		packet, err := tun.session.ReceivePacket()
		switch err {
		case nil:
			n := copy(buf[:], packet)
			tun.session.ReleaseReceivePacket(packet)
			tun.rate.update(uint64(n))
			return n, nil
		case windows.ERROR_NO_MORE_ITEMS:
			if !shouldSpin || uint64(nanotime()-start) >= spinloopDuration {
				windows.WaitForSingleObject(tun.readWait, windows.INFINITE)
				goto retry
			}
			procyield(1)
			continue
		case windows.ERROR_HANDLE_EOF:
			return 0, os.ErrClosed
		case windows.ERROR_INVALID_DATA:
			return 0, errors.New("Send ring corrupt")
		}
		return 0, fmt.Errorf("Read failed: %w", err)
	}
}

func (tun *NativeTun) Write(buf []byte) (int, error) {
	tun.running.Add(1)
	defer tun.running.Done()
	if tun.close.Load() {
		return 0, os.ErrClosed
	}

	packetSize := len(buf)
	tun.rate.update(uint64(packetSize))

	packet, err := tun.session.AllocateSendPacket(packetSize)
	switch err {
	case nil:
		// TODO: Explore options to eliminate this copy.
		n := copy(packet, buf[:])
		tun.session.SendPacket(packet)
		return n, nil
	case windows.ERROR_HANDLE_EOF:
		return 0, os.ErrClosed
	case windows.ERROR_BUFFER_OVERFLOW:
		return 0, nil // Dropping when ring is full.
	default:
		return 0, fmt.Errorf("Write failed: %w", err)
	}
}

// XXX: Use wintun, don't forgot add dll file
// ref: https://www.wintun.net/
// Create a new tun interface assign a ipv4 address and set link up
// addr: <ipv4 addr>/<netmask//
func SetupTun(addr *netlink.Addr) (*NativeTun, error) {
	suffix := utils.RandStr(4)
	ifname := "tun-" + suffix

	wt, err := wintun.CreateAdapter(ifname, WintunTunnelType, WintunStaticRequestedGUID)
	if err != nil {
		return nil, fmt.Errorf("Error creating interface: %w", err)
	}

	netmask := fmt.Sprintf("%d.%d.%d.%d", addr.IPNet.Mask[0], addr.IPNet.Mask[1], addr.IPNet.Mask[2], addr.IPNet.Mask[3])
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", fmt.Sprintf("name=%s", ifname), "static", addr.IP.String(), netmask)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("assign ip %s to %s %s", addr.IP.String(), ifname, err.Error())
	}

	tun := &NativeTun{
		wt:        wt,
		name:      ifname,
		handle:    windows.InvalidHandle,
		events:    make(chan Event, 10),
		forcedMTU: 1420,
	}

	tun.session, err = wt.StartSession(0x800000) // Ring capacity, 8 MiB
	if err != nil {
		tun.wt.Close()
		close(tun.events)
		return nil, fmt.Errorf("Error starting session: %w", err)
	}
	tun.readWait = tun.session.ReadWaitEvent()
	return tun, nil
}

func RoutiesAdd(iface string, routes []*net.IPNet, ip string) error {
	for _, rNet := range routes {
		netmask := fmt.Sprintf("%d.%d.%d.%d", rNet.Mask[0], rNet.Mask[1], rNet.Mask[2], rNet.Mask[3])
		cmd := exec.Command("route", "add", rNet.IP.String(), "mask", netmask, ip)

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
