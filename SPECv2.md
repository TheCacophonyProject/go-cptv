# Introduction

This document described the data format used to encode thermal video
recordings for the Cacophony Project. The format is designed to
efficiently represent thermal video data in a lossless way, and allow
for future extensibility.

# Data representation

The data format is binary (non-text). All numbers are represented in
little-endian form as this is the byte ordering used on the computer
platforms we use.

# Compression

The data format allows for compression of thermal video frames
internally and is then also compressed again using standard gzip
compression to obtain higher levels of compression. In order to read
data the entire file/stream must be passed through a gzip decompressor
first.

# Data Format

## Field Represention

Parts of the data format include variable length sections of
fields. To allow for the addition of extra fields in the future
without breaking existing readers, each field includes its length and
a identifying field code. Readers should skip over fields which they
don't recognise.

A field is represented as follows:

| Name   | Length | Type  | Description
| ------ | ------ | ----- |----------------------------------------------
| Length | 1      | uint8 | Length of the field's data
| Code   | 1      | char  | Character identifying
| Data   | ?      | ?     | Length & content depends on Length & Code

## Identification

The data format always starts with:

* 4 magic bytes: "CPTV"
* 1 byte: version code: 2

## Header

A single header should come next. It starts with:
* 1 byte indicating the header: "H"
* 1 byte indicating the number of fields in the header.

After these a number of fields will exist.

### Compulsory Header Fields

| Name          | Length   | Code  | Type    | Description
| ------------  | ------   | ----- | ------- | ---------------------------------------------
| Timestamp     | 8        | 'T'   | uint64  | Microseconds since 1970-01-01 UTC
| X resolution  | 4        | 'X'   | uint32  | Frame X resolution (columns)
| Y resolution  | 4        | 'Y'   | uint32  | Frame Y resolution (rows)
| Compression   | 1        | 'C'   | uint8   | Compression scheme in use (0 = uncompressed)
| Device name   | Variable | 'D'   | string  | Device name e.g. ("somewhere01")

### Optional header fields

| Name          | Length   | Code  | Type    | Description
| ------------  | ------   | ----- | ------- | ---------------------------------------------
| Motion config | Variable | 'M'   | string  | Motion detection configuration in YAML
| CameraSerial  | Variable | 'N'   | string  | Unique camera module serial number
| Model         | Variable | 'E'   | string  | Camera module model ("lepton3", "lepton3.5")
| Brand         | Variable | 'B'   | string  | Camera module brand ("flir")
| Firmware      | Variable | 'V'   | string  | Camera module firmware revision "{MAJOR}.{MINOR}.{BUILD}"
| DeviceID      | Variable | 'I'   | string  | Device id ("unique_id")
| Preview secs  | 1        | 'P'   | uint8   | Number of seconds of recording before motion event was detected
| Latitude      | 4        | 'L'   | float32 | Latitude of device location
| Longitude     | 4        | 'O'   | float32 | Longitude of device location
| LocTimestamp  | 8        | 'S'   | uint64  | Time at which location of device was set.  Microseconds since 1970-01-01 UTC
| Altitude      | 4        | 'A'   | float32 | Altitude of device location in metres.
| Accuracy      | 4        | 'U'   | float32 | Estimated accuracy of location settings in metres.
|BackgroundFrame| 1        | 'g'   | uint8 | Number of background frames in this file. In practise we are only checking this value is non zero, one should only expect a single background frame when reading a CPTV file

## Frames

One or more frames will follow the header. Each frame starts with:
* 1 byte indicating a frame: "F"
* 1 byte indicating the number of fields in the frame.

### Compulsory Frame Fields

The following frame fields must exist in every frame:

| Name          | Length | Code  | Type      | Description
| ----------    | ------ | ----- | --------- | ------------------------------------------------------------------
| Time on       | 4      | 't'   | uint32    | Time in ms since the camera was powered on
| Bit width     | 1      | 'w'   | uint8     | Bit width of the frame data
| Frame size    | 4      | 'f'   | uint32    | Size of the frame data
| Last FFC time | 4      | 'c'   | uint32    | Time of last Flat Field Correction (in ms since camera powered on)
| Last FFC TempC| 4      | 'b'   | float32   | Temperature at last FFC
| TempC         | 4      | 'a'   | float32   | Temperature of frame

### Optional Frame fields

| BackgroundFrame | 1        | 'g'   | uint8 | integer representation of a boolean 1 or 0 if this frame is a background frame


### Frame Data

Following the frame fields (as indicated by the field count at the
start of the frame) there will be a bytes of frame data. The number of
bytes will match the "frame size" header (code 'f') in the frame's fields.

Decoding the frame will involve use of the frame's bit width and the
compression scheme indicated in the header. For compression scheme 0
(no compression) read the frame data using the bit width
provided. Remember that data is always represented using little-endian
ordering.
