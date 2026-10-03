# Installs all pending software updates through the Windows Update Agent API
# (scripts/defender-vm.sh prepare). Runs as SYSTEM from a scheduled task: the
# API refuses to download or install from an ssh session. Progress goes to
# C:\wu\status.txt; the last line is "done <found> <installed> <reboot>" or
# "error <message>".
$ErrorActionPreference = 'Stop'
$status = 'C:\wu\status.txt'
function Say($m) { Add-Content -Path $status -Value $m }
try {
  Set-Content -Path $status -Value "start $(Get-Date -Format o)"
  $session = New-Object -ComObject Microsoft.Update.Session
  $result = $session.CreateUpdateSearcher().Search("IsInstalled=0 and IsHidden=0 and Type='Software'")
  $updates = New-Object -ComObject Microsoft.Update.UpdateColl
  foreach ($u in $result.Updates) {
    if ($u.InstallationBehavior.CanRequestUserInput) { Say "skip $($u.Title)"; continue }
    if (-not $u.EulaAccepted) { $u.AcceptEula() }
    Say "found $($u.Title)"
    [void]$updates.Add($u)
  }
  if ($updates.Count -eq 0) { Say "done 0 0 False"; exit 0 }
  $downloader = $session.CreateUpdateDownloader()
  $downloader.Updates = $updates
  [void]$downloader.Download()
  $installer = $session.CreateUpdateInstaller()
  $installer.Updates = $updates
  $r = $installer.Install()
  $ok = 0
  for ($i = 0; $i -lt $updates.Count; $i++) {
    # ResultCode 2 succeeded, 3 succeeded with errors, 4 failed, 5 aborted.
    $code = $r.GetUpdateResult($i).ResultCode
    if ($code -in 2, 3) { $ok++ }
    Say "result $code $($updates.Item($i).Title)"
  }
  Say "done $($updates.Count) $ok $($r.RebootRequired)"
} catch { Say ("error " + $_.Exception.Message) }
