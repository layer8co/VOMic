package ir.layer8.vomic

import android.Manifest
import android.content.Context
import android.media.AudioFormat
import android.media.AudioRecord
import android.media.MediaRecorder
import android.util.Log
import android.widget.Toast
import androidx.annotation.RequiresPermission
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import kotlinx.coroutines.yield
import okio.IOException
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetAddress
import java.net.Socket
import java.net.SocketTimeoutException
import java.nio.ByteBuffer

const val TAG = "AudioClient"

class AudioClient(
    private val context: Context,
    private val hostAddress: String,
    private val controlPort: Int,
    private val onStart: () -> Unit,
    private val onStop: (String) -> Unit,
) {
    private var audioRecord: AudioRecord? = null

    // socketsJob simply keeps track of sockets coroutine. Canceling it should be enough for halt everything.
    private var socketsJob: Job? = null


    // record_pcm will record from the device microphone and sends the bytes(pcm encoded) to the udp network socket.
    @RequiresPermission(Manifest.permission.RECORD_AUDIO)
    suspend fun start(parentScope: CoroutineScope) {

        val source = MediaRecorder.AudioSource.MIC
        val sample_rate = 44100
        val channelConfig = AudioFormat.CHANNEL_IN_MONO
        val audio_format = AudioFormat.ENCODING_PCM_16BIT
        val buffer_size = AudioRecord.getMinBufferSize(sample_rate, channelConfig, audio_format)

        // Instantiate the AudioRecorder. It has the ability to capture the audio is raw (PCM) format.
        // "PCM" is the most raw form of the computer digital sound system.
        // The Pipewire pipe-source is also expecting the sound data in this form (more specific, "PCM 16 bit little endian")
        // I don't know much about the internals of this codec. But it does the job.
        audioRecord = AudioRecord(
            source, sample_rate, channelConfig, audio_format, buffer_size
        )

        // Attempt to start recording
        // First check whether the audioRecord object is actually initialized
        if (audioRecord == null || audioRecord?.state != AudioRecord.STATE_INITIALIZED) {
            Log.e("device", "AudioRecord isn't initialized.")
            Toast.makeText(context, "Error when initializing the AudioRecord", Toast.LENGTH_SHORT)
                .show()
            return
        }

        // Now start the recording
        audioRecord!!.startRecording()

        onStart()

        // Control socket And SessionInit
        var mediaPort: UShort

        val port = withContext(Dispatchers.IO) {

            performTcpHandshake(
                hostAddress,
                controlPort,
                getSystemName(),
                parentScope,
                context,
                { stop(it) }
            )
        }


        mediaPort = port
        Log.i(TAG, "media port got: $mediaPort")
        if (mediaPort.toInt() == 0) {
            stop("Received port isn't look valid: $mediaPort")
            return
        }
        parentScope.launch {
            udpSendLoop(audioRecord!!, mediaPort, hostAddress)
        }
    }


    fun stop(cause: String = "Operation aborted") {
        Log.i(TAG, "$cause")

        onStop(cause)

        audioRecord?.apply {
            stop()
            release()
        }

        audioRecord = null

        socketsJob?.cancel()
        socketsJob = null
    }
}

// If it returns 0, it means something went wrong.
private suspend fun performTcpHandshake(
    host: String,
    port: Int,
    hostname: String,
    scope: CoroutineScope,
    context: Context,
    // A callback that halts everything in audioRecord class.
    OnTimeout: (String) -> Unit
): UShort = withContext(Dispatchers.IO) {
    var socket: Socket? = null
    try {
        socket = Socket(host, port)
        socket.soTimeout = 3000

        val init = SessionInit(1u, hostname)
        socket.outputStream.write(init.marshalBinary().array())

        val ackBuf = ByteArray(2)
        val read = socket.inputStream.read(ackBuf)
        if (read >= 2) {
            val ack = ByteBuffer.wrap(ackBuf).toSessionAck()
            Log.i(TAG, "Got media port: ${ack.media_port}")

            // Launch heartbeat func from a main-safe scope
            scope.launch {
                heartbeat(socket, OnTimeout)
            }

            return@withContext ack.media_port
        } else {
            Log.e(TAG, "TCP ack is abnormal: $read")
            return@withContext 0u
        }
    } catch (e: Exception) {
        Log.e(TAG, "TCP handshake failed: $e")
        withContext(Dispatchers.Main) {
            Toast.makeText(context, "Connection failed", Toast.LENGTH_SHORT).show()
        }
        return@withContext 0u
    }
}

// heartbeat will just listens and reads from the provided socket.
// It also cooperates with the coroutineScope and stops responding based on that.
private suspend fun heartbeat(
    socket: Socket,
    onTimeout: (String) -> Unit
) = withContext(Dispatchers.IO) {

    socket.soTimeout = 1000

    try {
        socket.getInputStream().use { input ->
            val buffer = ByteArray(1)

            while (isActive) {
                try {
                    val bytesRead = input.read(buffer)

                    Log.i(TAG, "Here got $bytesRead")

                    if (bytesRead == -1) {
                        onTimeout("Remote closed the connection")
                        break
                    }

                    yield()

                } catch (e: SocketTimeoutException) {
                    // Normal heartbeat: timeout just means no data arrived this cycle
                    if (!isActive) break
                    continue
                }
            }
        }
    } catch (e: IOException) {
        onTimeout("Socket error: ${e.message}")
    } finally {
        runCatching { socket.close() }
    }
}


private suspend fun udpSendLoop(
    ar: AudioRecord,
    mediaPort: UShort,
    remoteHost: String,
) = withContext(Dispatchers.IO) {
    val socket = DatagramSocket()
    val remote = InetAddress.getByName(remoteHost)
    val packetBuf = ByteArray(200)
    try {
        while (isActive) {
            val read = ar.read(packetBuf, 0, packetBuf.size)
            if (read > 0) {
                val packet = DatagramPacket(packetBuf, read, remote, mediaPort.toInt())
                socket.send(packet)
            }
            yield()
        }
    } finally {
        socket.close()
    }
}