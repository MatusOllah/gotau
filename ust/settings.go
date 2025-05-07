package ust

// Settings represents the settings of the UST file.
// It holds various metadata and configuration used in vocal synthesis.
type Settings struct {
	Tempo       float64 // Tempo is the tempo (in BPM).
	ProjectName string  // ProjectName is a human-readable name of the project.
	Project     string  // Project is the path to the project (used only in OpenUtau).
	VoiceDir    string  // VoiceDir is the directory path to the voicebank.
	OutFile     string  // OutFile is the path to the output audio file that will be generated.
	CacheDir    string  // CacheDir is the directory path for cached temporary data.
	Tool1       string  // Tool1 is the path to the first synthesis tool (e.g., wavtool).
	Tool2       string  // Tool2 is the path to the second synthesis tool (e.g., resampler).
	Mode2       bool    // Mode2 indicates whether Mode2 (advanced pitch editing) is enabled.
}
