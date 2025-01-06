export interface fetchFunc {
    (URL: string | URL, options?: RequestInit): Promise<Response>
}
export function makeIdFetch(id: number): fetchFunc {
    const func : fetchFunc = function (URL, {headers, ...options}={}) {
        return fetch(URL, {
            headers: {
                id: String(id),
                ...headers
            },
            ...options
        });
    }
    return func;
}
