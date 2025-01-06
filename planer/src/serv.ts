export type fetchFunc = (URL: string | URL, options: RequestInit) => Promise<Response>
export function makeIdFetch(id: number): fetchFunc {
    return function (URL: string | URL, {headers, ...options}: RequestInit): Promise<Response> {
        console.log({
            headers: {
                id: String(id),
                ...headers
            },
            ...options
        }, options);
        return fetch(URL, {
            headers: {
                id: String(id),
                ...headers
            },
            ...options
        });
    }
}
