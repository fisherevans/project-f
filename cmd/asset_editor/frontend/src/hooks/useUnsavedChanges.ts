import { useEffect } from "react";
import { useBlocker } from "react-router";

export function useUnsavedChanges(dirty: boolean) {
    const blocker = useBlocker(({ currentLocation, nextLocation }) => {
        return dirty && currentLocation.pathname !== nextLocation.pathname;
    });

    useEffect(() => {
        if (blocker.state === "blocked") {
            const proceed = window.confirm("You have unsaved changes. Leave this page?");
            if (proceed) {
                blocker.proceed();
            } else {
                blocker.reset();
            }
        }
    }, [blocker]);

    useEffect(() => {
        if (!dirty) return;
        const handler = (e: BeforeUnloadEvent) => {
            e.preventDefault();
        };
        window.addEventListener("beforeunload", handler);
        return () => window.removeEventListener("beforeunload", handler);
    }, [dirty]);
}
