package ir.layer8.vomic

import android.Manifest
import android.app.Activity
import android.content.pm.PackageManager
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.viewModels
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel
import ir.layer8.vomic.ui.theme.VAndroid_Theme
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.nio.ByteBuffer
import java.nio.ByteOrder


class MainActivity : ComponentActivity() {


    override fun onCreate(savedInstanceState: Bundle?) {

        val viewModel: AudioRecordViewModel by viewModels()

        super.onCreate(savedInstanceState)
        setContent {
            VAndroid_Theme {
                Scaffold(topBar = { MinimalDropdownAppBar() }) { padding ->
                    SimpleLogoWithButton(
                        modifier = Modifier.padding(padding),
                        viewModel = viewModel
                    )
                }
            }
        }
    }
}

///////////////////////////////////////////////////////////
// UI

// Newer android devices, Specially Xiaomi devices are resonated
// with their annoying optimization mechanism.
// This checkbox can help prevent the device to get into sleep.
@Composable
fun KeepTheScreenOnCheckBox() {
    var checked by remember { mutableStateOf(false) }

    Row(
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Checkbox(
            checked = checked,
            onCheckedChange = { checked = it }
        )
        Text(
            "Keep the screen on"
        )
    }

    val currentView = LocalView.current
    DisposableEffect(checked) {
        currentView.keepScreenOn = checked
        onDispose {
            currentView.keepScreenOn = false
        }
    }
}

@Composable
fun AboutDialog(onDismissRequest: () -> Unit) {
    Dialog(onDismissRequest = { onDismissRequest() }) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .height(200.dp)
                .padding(16.dp),
            shape = RoundedCornerShape(16.dp),
        ) {

            val versionCode = BuildConfig.VERSION_CODE
            Text(
                text = "VO Mic for Android \n \n $versionCode",
                modifier = Modifier
                    .fillMaxSize()
                    .wrapContentSize(Alignment.Center),
                textAlign = TextAlign.Center,
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MinimalDropdownAppBar() {
    var expanded by remember { mutableStateOf(false) }
    var showDialog by remember { mutableStateOf(false) }
    TopAppBar(
        title = { Text("VO Mic") },
        actions = {
            IconButton(onClick = { expanded = true }) {
                Icon(Icons.Default.MoreVert, contentDescription = "More options")
            }

            DropdownMenu(
                expanded = expanded,
                onDismissRequest = { expanded = false }
            ) {
                DropdownMenuItem(
                    text = { Text("About") },
                    onClick = {
                        expanded = false
                        showDialog = !showDialog
                    }
                )
            }
        }
    )

    if (showDialog) {
        AboutDialog {
            showDialog = !showDialog
        }
    }
}

@Composable
fun SimpleLogoWithButton(
    modifier: Modifier = Modifier,
    viewModel: AudioRecordViewModel = viewModel()
) {
    val context = LocalContext.current

    val prefStore = remember { VPreferences(context) }

    var hostAddress by remember { mutableStateOf("") }

    var portText by remember { mutableStateOf("") }

    val IsRecording by viewModel.RecordingState.collectAsState()

    // Load the previous Host & Port form the DataStore, Only the first time when user enters
    // single time, when the user started the app.
    LaunchedEffect(Unit) {
        hostAddress = prefStore.readHost()
        portText = prefStore.readPort()
    }

    val port = portText.toIntOrNull() ?: 0

    val audioClient = AudioClient(
        context.applicationContext,
        hostAddress,
        port,
        onStart = {
            viewModel.startRecordingState()
        },
        onStop = { message ->
            viewModel.stopRecordingState()  // This should change the state that the main button depends on.

            // The onStop Callback might be called on a different context than Main.
            // So Execute post this code on the Main Thread queue.
            Handler(Looper.getMainLooper()).post {
                Toast.makeText(context, message, Toast.LENGTH_SHORT).show()
            }
        }
    )

    // recording job can help us to cancel the entire
    // so we can halt the operation later by canceling that.
    var recordingJob by remember { mutableStateOf<Job?>(null) }

    // main scope that wil
    val scope = rememberCoroutineScope()

    // Auto-stop when leaving the screen
    DisposableEffect(Unit) {
        onDispose {
            recordingJob?.cancel()
            viewModel.stopRecordingState()
        }
    }

    val permissionRequestLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            // Thanks. Got the permission. Now press the button again.
        } else {
            Toast.makeText(context, "Microphone permission denied.", Toast.LENGTH_SHORT).show()
        }
    }

    Surface(
        modifier = modifier.fillMaxSize(),
        color = MaterialTheme.colorScheme.background
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.SpaceEvenly
        ) {
            Icon(
                painter = painterResource(id = R.drawable.microphone_image),
                contentDescription = "Logo",
                tint = MaterialTheme.colorScheme.onSurface,
                modifier = Modifier.size(130.dp),
            )
            Column {
                HostInput(
                    hostAddress = hostAddress,
                    onHostChange = { hostAddress = it }
                )
                Spacer(modifier = Modifier.height(20.dp))
                PortInput(
                    portAddress = portText,
                    onPortChange = { portText = it }
                )
                Spacer(modifier = Modifier.height(20.dp))
                KeepTheScreenOnCheckBox()
            }

            Button(
                onClick = {
                    if (IsRecording) {
                        recordingJob?.cancel()
                        recordingJob = null
                        audioClient.stop()
                    } else {
                        val hasPerm = ContextCompat.checkSelfPermission(
                            context, Manifest.permission.RECORD_AUDIO
                        ) == PackageManager.PERMISSION_GRANTED

                        val hasWifi = checkWifiOnAndConnected(context)

                        if (hasPerm && hasWifi) {
                            recordingJob = scope.launch {
                                prefStore.saveHost(hostAddress)
                                prefStore.savePort(portText)
                                audioClient.start(this)
                            }
                        }
                        when {
                            !hasPerm -> {
                                permissionRequestLauncher.launch(Manifest.permission.RECORD_AUDIO)
                            }

                            !hasWifi -> {
                                Toast.makeText(
                                    context,
                                    "Check your network connection.",
                                    Toast.LENGTH_SHORT
                                ).show()
                            }
                        }
                    }
                },
                colors = ButtonDefaults.buttonColors(
                    containerColor = if (IsRecording) MaterialTheme.colorScheme.error
                    else MaterialTheme.colorScheme.primary
                ),
                modifier = Modifier.padding(top = 8.dp)
            ) {
                Text(text = if (IsRecording) "Stop Recording" else "Connect")
            }


            if (IsRecording) {
                CircularProgressIndicator(modifier = Modifier.size(24.dp))
            }
        }
    }
}

@Composable
fun HostInput(
    hostAddress: String,
    onHostChange: (String) -> Unit
) {
    OutlinedTextField(
        value = hostAddress,
        onValueChange = onHostChange,
        label = { Text("Host Address") }
    )
}


@Composable
fun PortInput(
    portAddress: String,
    onPortChange: (String) -> Unit
) {

    OutlinedTextField(
        value = portAddress,
        onValueChange = onPortChange,
        keyboardOptions = KeyboardOptions.Default.copy(keyboardType = KeyboardType.Number),
        label = { Text("Control port") }
    )
}

@Preview(showBackground = true)
@Composable
fun PreviewSimpleLogoWithButton() {
    VAndroid_Theme {
        SimpleLogoWithButton()
    }
}
