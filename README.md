# mkmemkey.exe

creates a **volatile Windows registry key** (only exists in memory, doesn't persist a Windows reboot).

# usage

```
mkmemkey [HKLM|HKCU|HKCR|HKU|HKCC]\[KeyName]
```

exit codes:
- `0` = success
- `1` = error
- `2` = key already existed, no changes made

# example
```batch
REM # create the memkey
$ mkmemkey "HKCU\SOFTWARE\MyNewKey"

REM # and optionally also add values to it (using usual registry tools like regedit.exe or reg.exe)
$ reg add "HKCU\SOFTWARE\MyNewKey" /v "foo" /d "bar" /t REG_SZ

REM # query it
$ reg query "HKCU\SOFTWARE\MyNewKey"
HKEY_CURRENT_USER\SOFTWARE\MyNewKey:
foo    REG_SZ    bar

REM # and when we query after a reboot, it's all gone
$ reg query "HKCU\SOFTWARE\MyNewKey"
ERROR: The system was unable to find the specified registry key or value.
```

# compatibility
Any app (i.e. regedit.exe, reg.exe, etc) should be able to create Registry *Values* under a memkey and also delete memkeys. However, most apps can **not** create **Subkeys** below a memkey (as such subkeys need to be created as volatile keys, too). Instead, `memkey.exe` can be used to create volatile Subkeys.