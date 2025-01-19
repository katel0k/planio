import { createContext } from "react"

export const NAME_COOKIE_KEY: string = 'name';
export const ID_UNSET: number = -1;
export function getNameCookie(): number {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + NAME_COOKIE_KEY.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? parseInt(decodeURIComponent(matches[1])) : ID_UNSET;
}

const IdContext: React.Context<number> = createContext(ID_UNSET);
export default IdContext;

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
