package ir.layer8.vomic

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities.TRANSPORT_VPN
import android.net.NetworkCapabilities.TRANSPORT_WIFI
import android.os.Build


// Mainly used for the SessionInit Handshake.
fun getSystemName(): String {
    return "${Build.MODEL}"
}



fun checkWifiOnAndConnected(context: Context): Boolean {
    val connectivityManager = context.applicationContext.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager


    val network = connectivityManager.activeNetwork ?: return false

    val capabilities = connectivityManager.getNetworkCapabilities(network) ?: return false


    val status = capabilities.hasTransport(TRANSPORT_WIFI) || capabilities.hasTransport(TRANSPORT_VPN)
    return status
}
