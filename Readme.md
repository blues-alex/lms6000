[RU](Readme_ru.md)

The provided Go code defines a package `lms6000` that interfaces with an LMS6000 device via UART. The code includes constants, variables, types, and methods to communicate with the device, perform measurements, and parse data.

### Key Components

1. **Constants and Variables:**
   
   - Constants like `DATA_OFS`, `TIMEDATE_OFS`, and `DATA_LENGTH` define offsets for reading specific parts of the measurement data.
   - Command byte slices (`HALLO_MSG`, `START_MEASURE`, etc.) are used to send commands to the device.

2. **Types:**
   
   - `Measure`: A struct that holds measurement values and spectrum data in maps, which can be serialized to JSON.
   - `LMS6000`: Represents a connection to the LMS6000 device, including methods for communication and configuration.

3. **Methods:**
   
   - **`NewLMS6000(p string, b int) (*LMS6000, error)`**: Initializes a new connection to the LMS6000 device using UART.
   
   - **`GetTimeAverage()`**: Retrieves the current time averaging setting from the device.
   
   - **`SetTimeAverage(t time.Duration) error`**: Sets the time averaging duration for measurements on the device.
   
   - **`GetMeasure() (*Measure, error)`**: Initiates a measurement and retrieves data. It waits for the specified `TimeAverage` before reading the response.
   
   - **`SwitchToMultiMeasureMode()`**: Switches the device to multi-measurement mode.
   
   - **`StartMeasure() error`**: Starts a single measurement on the device.

4. **Helper Functions:**
   
   - **`bytesToFloats(buf []byte) []float32`**: Converts a byte slice into an array of `float32` values, interpreting each 4-byte segment as a little-endian float.
   
   - **`newMeasure() *Measure`**: Creates and returns a new `Measure` struct with initialized maps for values and spectrum data.
   
   - **`parseMeasure(m []byte) *Measure`**: Parses the raw byte response from the device into a structured `Measure` object, mapping measurement keys to their respective float values.

### Usage

To use this package:

1. **Initialize the Device:**
   
   ```go
   lms, err := lms6000.NewLMS6000("/dev/ttyUSB0", 9600)
   if err != nil {
       log.Fatal(err)
   }
   ```

2. **Set Time Average (Optional):**
   
   ```go
   err = lms.SetTimeAverage(100 * time.Millisecond)
   if err != nil {
       log.Fatal(err)
   }
   ```

3. **Get Measurement:**
   
   ```go
   measurement, err := lms.GetMeasure()
   if err != nil {
       log.Fatal(err)
   }
   fmt.Println(measurement.Values) // Access the measurement values
   fmt.Println(measurement.Spectrum) // Access the spectrum data
   ```

### Considerations

- Ensure that the UART connection is correctly configured and accessible.
- Handle errors appropriately, especially when dealing with hardware communication.
- The `TimeAverage` setting affects how long the device waits before sending a response; adjust it based on your measurement needs.

This package provides a structured way to interact with an LMS6000 device, abstracting the low-level UART communication details.

[Licanse Apache 2.0](LICENSE)