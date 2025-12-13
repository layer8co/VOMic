package ir.layer8.vomic

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

// Creating a datastore. Name is actually the name of the file that is used to persist the data.
class VPreferences(val context: Context) {
    private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "pref")

    private val HOST_ADDRESS = stringPreferencesKey("host_address")
    private val HOST_PORT = stringPreferencesKey("host_port")

    suspend fun readHost(): String {
        return context.dataStore.data.map { preferences ->
            preferences[HOST_ADDRESS] ?: ""
        }.first()
    }

    suspend fun saveHost(value: String) {
        context.dataStore.updateData { it ->
            it.toMutablePreferences().also {
                it[HOST_ADDRESS] = value
            }
        }
    }

    suspend fun readPort(): String {
        return context.dataStore.data.map { preferences ->
            preferences[HOST_PORT] ?: ""
        }.first()
    }

    suspend fun savePort(value: String) {
        context.dataStore.updateData { it ->
            it.toMutablePreferences().also {
                it[HOST_PORT] = value
            }
        }
    }

}
