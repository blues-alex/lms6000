## Project Overview [Readme_RU](./Readme_RU.md)

The `lms6000` project is a library designed to interact with the LMS6000 instrument, which measures spectral light characteristics. The library provides functionality to send commands to the instrument, retrieve measurement data, and configure parameters.

**Key Components:**

*   **`SerialDevice` Interface:** Defines the interface for interacting with a serial port. This allows for different serial port implementations (e.g., a physical serial port, a mock for testing).
*   **`LMS6000` Struct:** Represents the LMS6000 instrument and holds information about current settings (e.g., averaging time).
*   **Commands:** A set of constants representing byte sequences used to send commands to the instrument.
*   **`Meassure` Struct:** A structure for storing measurement results, including value maps and spectral data.

**Used Libraries and Frameworks:**

*   `encoding/binary`: For converting data between different formats (e.g., `uint32` to bytes).
*   `fmt`: For formatted input/output.
*   `math`: For mathematical operations, specifically for converting `float32`.
*   `time`: For working with time.
*   `github.com/blues-alex/clog`: For logging.

## File `lms6000.go`

### `package lms6000`

Declares the package containing the library's code.

### `import (...)`

Imports the necessary packages.

### `const (...)`

Defines constants used throughout the project.

*   `DATA_OFS`: Offset of the measurement data in the byte representation of the instrument's response.
*   `TIMEDATE_OFS`: Offset of the date and time information in the byte representation of the instrument's response.
*   `DATA_LENGTH`: Length of the measurement data in the byte representation of the instrument's response.
*   `HALLO_MESS`: Greeting command sent to the instrument.
*   `START_MEASSURE`: Command to initiate a single measurement.
*   `START_MULTI_MEASSURES`: Command to initiate measurements in multi-channel mode.
*   `GET_DATA`: Command to request measurement data.
*   `SET_TIME_AVERAVE`: Command to set the averaging time.
*   `GET_TIME_AVERAVE`: Command to request the current averaging time.
*   `MEASSURE_KEYS`: An array of strings representing the names of the measured parameters.

### `type Meassure struct`

Defines the `Meassure` structure, which holds measurement results.

*   `Values`: A map containing the values of the measured parameters (e.g., PAR, PPFD). Keys are parameter names, values are the measured values.
*   `Spectrum`: A map containing spectral data. Keys are spectral channel indices, values are the measured values.

### `type SerialDevice interface`

Defines the `SerialDevice` interface, which specifies methods for interacting with a serial port.

*   `Write([]byte) (int, error)`: Sends data to the serial port. Returns the number of bytes sent and an error (if any occurred).
*   `ReadMessage() ([]byte, error)`: Reads data from the serial port. Returns the read bytes and an error (if any occurred).
*   `Close() error`: Closes the connection to the serial port.

### `func NewLMS6000(device SerialDevice) (*LMS6000, error)`

Constructor for creating an `LMS6000` instance.

*   **Parameters:**
    *   `device`: An implementation of the `SerialDevice` interface.
*   **Return Values:**
    *   `*LMS6000`: A pointer to the created `LMS6000` instance.
    *   `error`: An error (if any occurred).
*   **Description:**
    *   Creates a new `LMS6000` instance, initializes its `device` and `TimeAverage`.
    *   Retrieves the current averaging time from the instrument and sets it within the `LMS6000` structure.

### `func (l *LMS6000) GetTimeAverage() (time.Duration, error)`

Retrieves the current averaging time from the instrument.

*   **Parameters:** None.
*   **Return Values:**
    *   `time.Duration`: The current averaging time.
    *   `error`: An error (if any occurred).
*   **Description:**
    *   Sends the `GET_TIME_AVERAVE` command to the instrument.
    *   Reads the instrument's response.
    *   Converts the received data into a `time.Duration`.

### `func (l *LMS6000) SetTimeAverage(t time.Duration) error`

Sets the averaging time on the instrument.

*   **Parameters:**
    *   `t`: The new averaging time.
*   **Return Values:**
    *   `error`: An error (if any occurred).
*   **Description:**
    *   Converts the `time.Duration` to a `uint32`.
    *   Sends the `SET_TIME_AVERAVE` command to the instrument, along with the new averaging time data.
    *   Reads the instrument's response.

### `func (l *LMS6000) GetMeassure() (*Meassure, error)`

Performs a measurement and returns the results.

*   **Parameters:** None.
*   **Return Values:**
    *   `*Meassure`: A structure containing the measurement results.
    *   `error`: An error (if any occurred).
*   **Description:**
    *   Initiates a measurement using `StartMeassure`.
    *   Pauses for the averaging time.
    *   Sends the `GET_DATA` command to the instrument.
    *   Reads the instrument's response.
    *   Parses the received data and creates a `Meassure` structure.

### `func (l *LMS6000) SwithToMultiMeassureMod() error`

Switches the instrument to multi-channel measurement mode.

*   **Parameters:** None.
*   **Return Values:**
    *   `error`: An error (if any occurred).
*   **Description:**
    *   Sends the `START_MULTI_MEASSURES` command to the instrument.
    *   Reads the instrument's response.

### `func (l *LMS6000) StartMeassure() error`

Initiates a single measurement.

*   **Parameters:** None.
*   **Return Values:**
    *   `error`: An error (if any occurred).
*   **Description:**
    *   Sends the `START_MEASSURE` command to the instrument.
    *   Reads the instrument's response.

### `func bytesToFloats(buf []byte) []float32`

Converts a byte array to a `float32` array.

*   **Parameters:**
    *   `buf`: The byte array.
*   **Return Values:**
    *   `[]float32`: The `float32` array.
*   **Description:**
    *   Converts each set of 4 bytes into a `float32`.

### `func newMeassure() *Meassure`

Creates a new instance of the `Meassure` structure.

*   **Parameters:** None.
*   **Return Values:**
    *   `*Meassure`: A pointer to the created `Meassure` instance.
*   **Description:**
    *   Initializes the `Values` and `Spectrum` maps within the `Meassure` structure.

### `func parseMeassure(m []byte) *Meassure`

Parses a byte array received from the instrument and creates a `Meassure` structure.

*   **Parameters:**
    *   `m`: The byte array received from the instrument.
*   **Return Values:**
    *   `*Meassure`: A structure containing the measurement results.
*   **Description:**
    *   Extracts measurement data and spectral data from the byte array.
    *   Populates the `Values` and `Spectrum` maps within the `Meassure` structure.

## Example Usage

```go
// Example of using the LMS6000 library

// Assume you have an implementation of SerialDevice, such as MySerialPort
myPort := &MySerialPort{ /* ... your serial port configuration ... */ }

// Create an LMS6000 instance
lms, err := NewLMS6000(myPort)
if err != nil {
  // Handle the error
}

// Retrieve measurement data
measure, err := lms.GetMeassure()
if err != nil {
  // Handle the error
}

// Display the results
fmt.Println(measure.Values)
fmt.Println(measure.Spectrum)

// Set the averaging time
err = lms.SetTimeAverage(10 * time.Second)
if err != nil {
  // Handle the error
}
```
# [LICENSE](./LICENSE)
