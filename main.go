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
)

const MTU = 800

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
  run("ip", "addr", "add", "10.8.8.1/24", "dev", name)
  run("ip", "link", "set", "dev", name, "mtu", strconv.Itoa(MTU))

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
  buffer := make([]byte, MTU)
  var count int
  var err error
  start := time.Now()
  for {
    count, err = file.Read(buffer)
    if err != nil {
      break
    }
    fmt.Printf("%v %v\n",
      time.Now().Sub(start),
      count)
  }
}
