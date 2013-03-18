package main

/*
#include "tun.c"
*/
import "C"
import (
  "unsafe"
  "fmt"
  "log"
  "os/exec"
  "strings"
  "os"
  "strconv"
  "time"
  "flag"
  "net"
)

const MTU = 800

var ip = flag.String("ip", "10.8.8.1", "ip for tun device")
var listen = flag.String("listen", ":39876", "listen address")
var remote = flag.String("remote", "none", "remote address")
var gateway = flag.Bool("gateway", false, "set as default gateway")

func init() {
  flag.Parse()
}

func NewTun() (fd C.int, name string) {
  tun_name := C.make_empty_name();
  defer C.free(unsafe.Pointer(tun_name))
  fd = C.tun_alloc(tun_name, C.IFF_TUN)
  if fd < C.int(0) {
    log.Fatal("cannot create tun device")
  }
  name = C.GoString(tun_name)

  fmt.Printf("fd %d, name %s\n", fd, name)
  run("ip", "link", "set", name, "up")
  run("ip", "addr", "add", *ip + "/24", "dev", name)
  run("ip", "link", "set", "dev", name, "mtu", strconv.Itoa(MTU))
  if *gateway {
    split := strings.Split(*ip, ".")
    split[len(split) - 1] = "0"
    network := strings.Join(split, ".")
    run("ip", "route", "change", network + "/24", "via", *ip)
  }

  return fd, name
}

func run(cmd string, args ...string) {
  out, err := exec.Command(cmd, args...).Output()
  if err != nil {
    log.Fatalf("error on running command %s %s\n>> %s <<",
      cmd,
      strings.Join(args, " "),
      out)
  }
}

func main() {
  fd, name := NewTun()
  file := os.NewFile(uintptr(fd), name)
  remotes := make(map[string]*net.UDPAddr)
  start := time.Now()

  addr, err := net.ResolveUDPAddr("udp", *listen)
  if err != nil {
    log.Fatal("invalid listen address")
  }
  conn, err := net.ListenUDP("udp", addr)
  if err != nil {
    log.Fatal("fail to listen udp")
  }
  go func() {
    buffer := make([]byte, MTU * 2)
    var count int
    var err error
    var remoteAddr *net.UDPAddr
    fmt.Printf("listening %v\n", addr)
    for {
      count, remoteAddr, err = conn.ReadFromUDP(buffer)
      fmt.Printf("%v read from udp %v %v\n",
        time.Now().Sub(start),
        remoteAddr,
        count)
      if err != nil {
        break
      }
      if remotes[remoteAddr.String()] == nil {
        remotes[remoteAddr.String()] = remoteAddr
      }
      fmt.Printf("write to tun %d\n", count)
      file.Write(buffer[:count])
    }
  }()

  if *remote != "none" {
    remoteAddr, err := net.ResolveUDPAddr("udp", *remote)
    if err != nil {
      log.Fatal("invalid remote address")
    }
    remotes[remoteAddr.String()] = remoteAddr
  }

  buffer := make([]byte, MTU * 2)
  var count int
  for {
    count, err = file.Read(buffer)
    fmt.Printf("%v read from tun %v\n",
      time.Now().Sub(start),
      count)
    if err != nil {
      break
    }
    for _, remoteAddr := range(remotes) {
      fmt.Printf("write to %v\n", remoteAddr)
      conn.WriteToUDP(buffer[:count], remoteAddr)
    }
  }
}
