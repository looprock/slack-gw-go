# slack-gw

slack-gw serves two functions:
- provide a common API for all your internal (RE: inside a vpc, etc) appications to submit to without needing to embed slack tokens everywhere
- provide the ability to send to more than one channel at the same time

# environment variables

**SLACKTOKEN** - (required) the slack token to use for your gateway

**GWPORT** - (optional - default: 8080) the port for the gateway to listen on

# API

slack-gw only supports **POST** requests of JSON to /

## JSON object

**message** - (required) the content of the message

**channels** - (required) a list of channels you want to send the message to

**topic** - (optional) If you add a topic, it will get prepended to the message in the format: 'topic - message'

## Example

```
{
    "message": "This is a message you want to send", 
    "channels": [
        "channel1", 
        "channel2"], 
    "topic": "MYSCRIPT"
}
```

## Test

```
SLACKTOKEN=sometokenhere GWPORT=8082 go run slack-gw.go

curl -vX POST http://localhost:8082 -d @input.json --header "Content-Type: application/json"
```
