# Doorbell

Code for a raspberry pi sitting near the door of my building that detects presses of the doorbell and sends notifications to the various units.

This is necessary because wireless doorbells can't reach some of the other units (whereas I'm on the first floor so can detect the doorbell presses better). Also it was fun to setup :)

## Setup

### Hardware

- Raspberry pi sitting near the front door running, running the controller code.
- Door bells by the front door, speaking the zigbee protocol
- My phone, which gets notifications when the bell is rung


### SaaS

- https://healthchecks.io/ the controller hits this site every 10m, so I can get an email if it goes down (or reports an issue)
- http://ntfy.sh/ this is where I send notification messages that show up on my phone


### Software

- [Raspberry PI OS](https://www.raspberrypi.com/software/) running the OS on the Raspberry PI
- [age](https://github.com/FiloSottile/age) for encrypting sensitive configuration 
- [zigbee2mqtt](https://www.zigbee2mqtt.io/) for translating the zigbee communication
- [mosquitto](https://mosquitto.org/) as a broker to get messages from zigbee

## Developing

### Configuration

The configuration lives in `conf/config.yaml`. Under `units` it lists the units corresponding to the doorbells, physical specs about the doorbells, and what happens when they are rung.

If `ntfy` is set, then a message is sent via [http://ntfy.sh/](http://ntfy.sh/). The topics on ntfy are all public, so the convention is to put a nonce in the URL. Hence we encrypt the URL.

If `chime` is set, then a message is sent via zigbee to the address listed. Right now it's hard coded to send `true` on `set/alarm` because that's what my current chime accepts, but we can change that.

### Deployment

The `./release.sh` script builds the go binary and scp's it to the doorbell on the local network. Optionally you can `--restart`.
