#ifndef _HWISENSSM2_H_INCLUDED_
#define _HWISENSSM2_H_INCLUDED_

// Name of the file mapping object that needs to be opened using OpenFileMapping Function:
#define HWiNFO_SENSORS_MAP_FILE_NAME2 "Global\\HWiNFO_SENS_SM2"

// Name of the global mutex which is acquired when accessing the Shared Memory space. Release as quick as possible !
#define HWiNFO_SENSORS_SM2_MUTEX "Global\\HWiNFO_SM2_MUTEX"

#define HWiNFO_SENSORS_STRING_LEN2 128
#define HWiNFO_UNIT_STRING_LEN 16

enum SENSOR_READING_TYPE
{
  SENSOR_TYPE_NONE = 0,
  SENSOR_TYPE_TEMP,
  SENSOR_TYPE_VOLT,
  SENSOR_TYPE_FAN,
  SENSOR_TYPE_CURRENT,
  SENSOR_TYPE_POWER,
  SENSOR_TYPE_CLOCK,
  SENSOR_TYPE_USAGE,
  SENSOR_TYPE_OTHER
};
typedef enum SENSOR_READING_TYPE SENSOR_READING_TYPE;

// No alignment of structure members
#pragma pack(1)

typedef struct _HWiNFO_SENSORS_READING_ELEMENT
{

  SENSOR_READING_TYPE tReading;                 // Type of sensor reading
  DWORD dwSensorIndex;                          // This is the index of sensor in the Sensors[] array to which this reading belongs to
  DWORD dwReadingID;                            // A unique ID of the reading within a particular sensor
  char szLabelOrig[HWiNFO_SENSORS_STRING_LEN2]; // Original label (e.g. "Chassis2 Fan")
  char szLabelUser[HWiNFO_SENSORS_STRING_LEN2]; // Label displayed, which might have been renamed by user
  char szUnit[HWiNFO_UNIT_STRING_LEN];          // e.g. "RPM"
  double Value;
  double ValueMin;
  double ValueMax;
  double ValueAvg;

} HWiNFO_SENSORS_READING_ELEMENT, *PHWiNFO_SENSORS_READING_ELEMENT;

typedef struct _HWiNFO_SENSORS_SENSOR_ELEMENT
{

  DWORD dwSensorID;                                  // A unique Sensor ID
  DWORD dwSensorInst;                                // The instance of the sensor (together with dwSensorID forms a unique ID)
  char szSensorNameOrig[HWiNFO_SENSORS_STRING_LEN2]; // Original sensor name
  char szSensorNameUser[HWiNFO_SENSORS_STRING_LEN2]; // Sensor name displayed, which might have been renamed by user

} HWiNFO_SENSORS_SENSOR_ELEMENT, *PHWiNFO_SENSORS_SENSOR_ELEMENT;

typedef struct _HWiNFO_SENSORS_SHARED_MEM2
{

  DWORD dwSignature;    // "HWiS" if active, 'DEAD' when inactive
  DWORD dwVersion;      // v1 is latest
  DWORD dwRevision;     //
  __time64_t poll_time; // last polling time

  // descriptors for the Sensors section
  DWORD dwOffsetOfSensorSection; // Offset of the Sensor section from beginning of HWiNFO_SENSORS_SHARED_MEM2
  DWORD dwSizeOfSensorElement;   // Size of each sensor element = sizeof( HWiNFO_SENSORS_SENSOR_ELEMENT )
  DWORD dwNumSensorElements;     // Number of sensor elements

  // descriptors for the Readings section
  DWORD dwOffsetOfReadingSection; // Offset of the Reading section from beginning of HWiNFO_SENSORS_SHARED_MEM2
  DWORD dwSizeOfReadingElement;   // Size of each Reading element = sizeof( HWiNFO_SENSORS_READING_ELEMENT )
  DWORD dwNumReadingElements;     // Number of Reading elements

} HWiNFO_SENSORS_SHARED_MEM2, *PHWiNFO_SENSORS_SHARED_MEM2;

#pragma pack()

#endif

// ***************************************************************************************************************
//                                          HWiNFO Shared Memory Footprint
// ***************************************************************************************************************
//
//         |-----------------------------|-----------------------------------|-----------------------------------|
// Content |  HWiNFO_SENSORS_SHARED_MEM2 |  HWiNFO_SENSORS_SENSOR_ELEMENT[]  | HWiNFO_SENSORS_READING_ELEMENT[]  |
//         |-----------------------------|-----------------------------------|-----------------------------------|
// Pointer |<--0                         |<--dwOffsetOfSensorSection         |<--dwOffsetOfReadingSection        |
//         |-----------------------------|-----------------------------------|-----------------------------------|
// Size    |  dwOffsetOfSensorSection    |   dwSizeOfSensorElement           |    dwSizeOfReadingElement         |
//         |                             |      * dwNumSensorElement         |       * dwNumReadingElement       |
//         |-----------------------------|-----------------------------------|-----------------------------------|
//
// ***************************************************************************************************************
//                                                  Code Example
// ***************************************************************************************************************
/*

HANDLE hHWiNFOMemory = OpenFileMapping( FILE_MAP_READ, FALSE, HWiNFO_SENSORS_MAP_FILE_NAME2 );
if (hHWiNFOMemory)
	PHWiNFO_SENSORS_SHARED_MEM2 pHWiNFOMemory = 
		(PHWiNFO_SENSORS_SHARED_MEM2) MapViewOfFile( hHWiNFOMemory, FILE_MAP_READ, 0, 0, 0 );
	
	// TODO: process signature, version, revision and poll time
	
	// loop through all available sensors
	for (DWORD dwSensor = 0; dwSensor < pHWiNFOMemory->dwNumSensorElements; dwSensor++)
	{
		PHWiNFO_SENSORS_SENSOR_ELEMENT sensor = (PHWiNFO_SENSORS_SENSOR_ELEMENT) ((BYTE*)pHWiNFOMemory + 
			pHWiNFOMemory->dwOffsetOfSensorSection + 
			(pHWiNFOMemory->dwSizeOfSensorElement * dwSensor));
		
		// TODO: process sensor
	}
	
	// loop through all available readings
	for (DWORD dwReading = 0; dwReading < pHWiNFOMemory->dwNumReadingElements; dwReading++)
	{
		PHWiNFO_SENSORS_READING_ELEMENT reading = (PHWiNFO_SENSORS_READING_ELEMENT) ((BYTE*)pHWiNFOMemory + 
			pHWiNFOMemory->dwOffsetOfReadingSection + 
			(pHWiNFOMemory->dwSizeOfReadingElement * dwReading));
		
		// TODO: process reading
	}
}

*/