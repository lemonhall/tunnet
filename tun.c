#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <unistd.h>
#include <string.h>
#include <netinet/in.h>
#include <linux/if.h>
#include <linux/if_tun.h>

int tun_alloc(char *dev, int flags) {
  struct ifreq ifr;
  int fd, err;
  char *clonedev = "/dev/net/tun";
  if ((fd = open(clonedev, O_RDWR)) < 0) {
    printf("cannot open clone device\n");
    return fd;
  }
  memset(&ifr, 0, sizeof(ifr));
  ifr.ifr_flags = flags;
  if (*dev) {
    strncpy(ifr.ifr_name, dev, IFNAMSIZ);
  }
  if ((err = ioctl(fd, TUNSETIFF, (void *)&ifr)) < 0) {
    close(fd);
    printf("cannot create tun device\n");
    return err;
  }
  strcpy(dev, ifr.ifr_name);
  return fd;
}

char* make_empty_name() {
  char *name = malloc(IFNAMSIZ);
  name[0] = '\0';
  return name;
}
