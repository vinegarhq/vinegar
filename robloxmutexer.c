#include <stdio.h>
#include <windows.h>

int
main() {
	HANDLE mutex;
	DWORD result;

	mutex = CreateMutexW(NULL, FALSE, L"ROBLOX_singletonMutex");

	if (!mutex) {
		fprintf(stderr, "failed to create mutex: %d\n", GetLastError());
		return 1;
	}

	result = WaitForSingleObject(mutex, 0);

	if (result == WAIT_TIMEOUT) {
		fprintf(stderr, "roblox mutex is already locked\n");
		CloseHandle(mutex);
		return 0;
	} else if (result != WAIT_OBJECT_0) {
		fprintf(stderr, "failed to grab roblox mutex: %d\n", GetLastError());
		return 1;
	}

	printf("roblox mutex locked\n");

	for (;;) {
		/*
		 * you cannot do anything until you kill me;
		 * by then, the mutex is yours.
		 */
	}

    return 0;
}
