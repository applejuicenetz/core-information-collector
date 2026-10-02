#define UNICODE
#define _UNICODE
#include <windows.h>
#include <msi.h>
#include <msiquery.h>
#include <strsafe.h>
#include <string.h>
#include <wchar.h>

#define PATH_CAPACITY 32768

static BOOL has_legacy_java(const WCHAR *directory) {
    WCHAR marker[PATH_CAPACITY];
    DWORD attributes;
    if (!directory[0] || FAILED(StringCchPrintfW(marker, PATH_CAPACITY, L"%s\\Java", directory))) {
        return FALSE;
    }
    attributes = GetFileAttributesW(marker);
    return attributes != INVALID_FILE_ATTRIBUTES && (attributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
}

static BOOL property_path(MSIHANDLE session, const WCHAR *property, WCHAR *path) {
    DWORD capacity = PATH_CAPACITY;
    path[0] = L'\0';
    return MsiGetPropertyW(session, property, path, &capacity) == ERROR_SUCCESS && path[0];
}

static BOOL registered_legacy(HKEY root, REGSAM view, const WCHAR *product, WCHAR *path) {
    HKEY uninstall;
    DWORD index;
    if (RegOpenKeyExW(root, L"Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall",
                     0, KEY_READ | view, &uninstall) != ERROR_SUCCESS) {
        return FALSE;
    }
    for (index = 0; ; index++) {
        WCHAR name[256];
        WCHAR display_name[512];
        DWORD name_size = 256;
        DWORD display_size = sizeof(display_name);
        DWORD path_size = PATH_CAPACITY * sizeof(WCHAR);
        HKEY entry;
        LSTATUS result = RegEnumKeyExW(uninstall, index, name, &name_size, NULL, NULL, NULL, NULL);
        if (result == ERROR_NO_MORE_ITEMS) break;
        if (result != ERROR_SUCCESS) continue;
        if (RegOpenKeyExW(uninstall, name, 0, KEY_READ | view, &entry) != ERROR_SUCCESS) continue;
        result = RegGetValueW(entry, NULL, L"DisplayName", RRF_RT_REG_SZ,
                              NULL, display_name, &display_size);
        if (result == ERROR_SUCCESS && _wcsnicmp(display_name, product, wcslen(product)) == 0 &&
            (display_name[wcslen(product)] == L'\0' || display_name[wcslen(product)] == L' ')) {
            result = RegGetValueW(entry, NULL, L"InstallLocation", RRF_RT_REG_SZ,
                                  NULL, path, &path_size);
            if (result == ERROR_SUCCESS) {
                size_t length = wcslen(path);
                if (length >= 2 && path[0] == L'"' && path[length - 1] == L'"') {
                    path[length - 1] = L'\0';
                    memmove(path, path + 1, (length - 1) * sizeof(WCHAR));
                }
                if (has_legacy_java(path)) {
                    RegCloseKey(entry);
                    RegCloseKey(uninstall);
                    return TRUE;
                }
            }
        }
        RegCloseKey(entry);
    }
    RegCloseKey(uninstall);
    return FALSE;
}

static BOOL find_legacy(MSIHANDLE session, WCHAR *path) {
    WCHAR folder[128];
    WCHAR product[128];
    WCHAR base[PATH_CAPACITY];
    const WCHAR *properties[] = {L"ProgramFiles64Folder", L"ProgramFilesFolder"};
    const WCHAR *variables[] = {L"ProgramW6432", L"ProgramFiles", L"ProgramFiles(x86)"};
    size_t index;
    DWORD capacity = 128;
    if (property_path(session, L"INSTALLDIR", path) && has_legacy_java(path)) return TRUE;
    if (MsiGetPropertyW(session, L"AJ_LEGACY_FOLDER", folder, &capacity) != ERROR_SUCCESS) return FALSE;
    capacity = 128;
    if (MsiGetPropertyW(session, L"AJ_LEGACY_PRODUCT", product, &capacity) != ERROR_SUCCESS) return FALSE;
    for (index = 0; index < sizeof(properties) / sizeof(properties[0]); index++) {
        if (property_path(session, properties[index], base) &&
            SUCCEEDED(StringCchPrintfW(path, PATH_CAPACITY, L"%sappleJuiceNETZ\\%s", base, folder)) &&
            has_legacy_java(path)) return TRUE;
    }
    for (index = 0; index < sizeof(variables) / sizeof(variables[0]); index++) {
        DWORD length = GetEnvironmentVariableW(variables[index], base, PATH_CAPACITY);
        if (length > 0 && length < PATH_CAPACITY &&
            SUCCEEDED(StringCchPrintfW(path, PATH_CAPACITY, L"%s\\appleJuiceNETZ\\%s", base, folder)) &&
            has_legacy_java(path)) return TRUE;
    }
    return registered_legacy(HKEY_LOCAL_MACHINE, KEY_WOW64_64KEY, product, path) ||
           registered_legacy(HKEY_LOCAL_MACHINE, KEY_WOW64_32KEY, product, path) ||
           registered_legacy(HKEY_CURRENT_USER, KEY_WOW64_64KEY, product, path) ||
           registered_legacy(HKEY_CURRENT_USER, KEY_WOW64_32KEY, product, path);
}

__declspec(dllexport) UINT __stdcall CheckLegacyInstallation(MSIHANDLE session) {
    WCHAR path[PATH_CAPACITY];
    WCHAR message[PATH_CAPACITY + 256];
    MSIHANDLE record;
    if (!find_legacy(session, path)) return ERROR_SUCCESS;
    StringCchPrintfW(message, PATH_CAPACITY + 256,
        L"Eine alte NSIS-Installation wurde in \"%s\" gefunden (Unterordner Java). "
        L"Bitte deinstallieren Sie zuerst das alte Setup und starten Sie danach dieses Setup erneut. "
        L"Die Installation wird abgebrochen.", path);
    record = MsiCreateRecord(1);
    MsiRecordSetStringW(record, 0, L"[1]");
    MsiRecordSetStringW(record, 1, message);
    MsiProcessMessage(session, INSTALLMESSAGE_ERROR | MB_OK | MB_ICONERROR, record);
    MsiCloseHandle(record);
    return ERROR_INSTALL_FAILURE;
}

#ifdef AJ_GUARD_TEST
int wmain(int argc, WCHAR **argv) {
    if (argc != 2) return 2;
    return has_legacy_java(argv[1]) ? 1 : 0;
}
#endif
