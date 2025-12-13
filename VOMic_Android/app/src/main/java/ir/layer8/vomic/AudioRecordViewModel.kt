package ir.layer8.vomic

import androidx.lifecycle.ViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

class AudioRecordViewModel : ViewModel() {

    private val _recordingState = MutableStateFlow(false)
    var RecordingState: StateFlow<Boolean> = _recordingState


    fun stopRecordingState() {
        _recordingState.value = false
    }

    fun startRecordingState() {
        _recordingState.value = true
    }
}