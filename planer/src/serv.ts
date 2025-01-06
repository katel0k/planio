export type fetchFunc = (URL: string | URL, options: RequestInit) => Promise<Response>
export function makeIdFetch(id: number): fetchFunc {
    return function (URL: string | URL, {headers, ...options}: RequestInit): Promise<Response> {
        return fetch(URL, {
            headers: {
                id: String(id),
                ...headers
            },
            ...options
        });
    }
}
