#!/usr/bin/env bash

qmp_powerdown() {
  echo '{ "execute": "qmp_capabilities" } { "execute": "system_powerdown" }';
  sleep 0.1;
}

qmp_powerdown | telnet localhost 40028

# #!/usr/bin/expect -f
# spawn telnet localhost 40028
# send "{ \"execute\": \"qmp_capabilities\" }\r"
# expect "{\"return\": {}}"
# send "{ \"execute\": \"system_powerdown\" }\r"
# expect "{\"return\": {}}"
