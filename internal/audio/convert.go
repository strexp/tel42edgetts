package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"

	"github.com/hajimehoshi/go-mp3"
)

func ConvertToWav(filePath string, mp3Data []byte, targetRate int) error {
	samples, err := resampleMp3ToPCM(mp3Data, targetRate)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	writeWavHeader(f, targetRate, len(samples))

	outBuf := make([]byte, 2)
	for _, sample := range samples {
		binary.LittleEndian.PutUint16(outBuf, uint16(sample))
		if _, err = f.Write(outBuf); err != nil {
			return err
		}
	}

	return nil
}

func resampleMp3ToPCM(mp3Data []byte, targetRate int) ([]int16, error) {
	dec, err := mp3.NewDecoder(bytes.NewReader(mp3Data))
	if err != nil {
		return nil, err
	}

	srcRate := dec.SampleRate()
	var monoSamples []int16
	buf := make([]byte, 4)

	for {
		_, err := io.ReadFull(dec, buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, err
		}

		left := int16(binary.LittleEndian.Uint16(buf[0:2]))
		right := int16(binary.LittleEndian.Uint16(buf[2:4]))
		mono := int16((int32(left) + int32(right)) / 2)
		monoSamples = append(monoSamples, mono)
	}

	srcLen := len(monoSamples)
	if srcLen == 0 {
		return []int16{}, nil
	}

	targetLen := int(float64(srcLen) * float64(targetRate) / float64(srcRate))
	targetSamples := make([]int16, targetLen)

	for i := range targetLen {
		srcPos := float64(i) * float64(srcRate) / float64(targetRate)
		index := int(srcPos)
		frac := srcPos - float64(index)

		if index >= srcLen-1 {
			targetSamples[i] = monoSamples[srcLen-1]
		} else {
			val1 := float64(monoSamples[index])
			val2 := float64(monoSamples[index+1])
			targetSamples[i] = int16(val1 + frac*(val2-val1))
		}
	}

	return targetSamples, nil
}

func writeWavHeader(w io.Writer, sampleRate int, numSamples int) {
	bitsPerSample, numChannels := 16, 1
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := numSamples * numChannels * bitsPerSample / 8
	chunkSize := 36 + dataSize

	header := make([]byte, 44)
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], uint32(chunkSize))
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16)
	binary.LittleEndian.PutUint16(header[20:22], 1)
	binary.LittleEndian.PutUint16(header[22:24], uint16(numChannels))
	binary.LittleEndian.PutUint32(header[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(header[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(header[34:36], uint16(bitsPerSample))
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataSize))

	_, _ = w.Write(header)
}
