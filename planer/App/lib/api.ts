import { createContext } from "react"
import { plan as planPB } from 'plan.proto'
import { auth as authPB } from 'auth.proto'

export const NAME_COOKIE_KEY: string = 'name';
export const ID_UNSET: number = -1;
export function getNameCookie(): number {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + NAME_COOKIE_KEY.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? parseInt(decodeURIComponent(matches[1])) : ID_UNSET;
}

export const IdContext: React.Context<number> = createContext(ID_UNSET);

export async function handleAuth(req: authPB.AuthRequest): Promise<authPB.AuthResponse> {
    const response: Response = await fetch(new URL("/auth", serverUrl), {
            method: "POST",
            headers: {
                'Content-Type': 'application/json;charset=utf-8'
            },
            body: JSON.stringify(req.toJSON()),
        })
    const buf: ArrayBuffer = await response.arrayBuffer();
    return authPB.AuthResponse.decode(new Uint8Array(buf));
}

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

export const serverUrl: URL = new URL("http://0.0.0.0:5000");

export interface API {
    getPlans: (options?: RequestInit) => Promise<planPB.UserPlans>,
    createPlan: (plan: planPB.NewPlanRequest, options?: RequestInit) => Promise<planPB.Plan>,
    changePlan: (change: planPB.ChangePlanRequest, options?: RequestInit) => Promise<planPB.Plan>,
    deletePlan: (del: planPB.DeletePlanRequest, options?: RequestInit) => Promise<planPB.Plan>
}

export function apiFactory(id: number): API {
    const url: URL = new URL("/plan", serverUrl);
    const f: fetchFunc = makeIdFetch(id);
    return {
        async getPlans(options?: RequestInit): Promise<planPB.UserPlans> {
            const response = await f(url, options);
            const buffer = await response.arrayBuffer();
            return planPB.UserPlans.decode(new Uint8Array(buffer));
        },
        async createPlan(plan: planPB.NewPlanRequest, options?: RequestInit): Promise<planPB.Plan> {
            const response = await f(url, {
                method: 'POST',
                body: planPB.NewPlanRequest.encode(plan).finish(),
                ...options,
            });
            const buffer = await response.arrayBuffer();
            return planPB.Plan.decode(new Uint8Array(buffer));
        },
        async changePlan(change: planPB.ChangePlanRequest, options?: RequestInit): Promise<planPB.Plan> {
            const response = await f(url, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json;charset=utf-8'
                },
                body: JSON.stringify(change.toJSON()),
                ...options,
            });
            const buffer = await response.arrayBuffer();
            return planPB.Plan.decode(new Uint8Array(buffer));
        },
        async deletePlan(del: planPB.DeletePlanRequest, options?: RequestInit): Promise<planPB.Plan> {
            const response = await f(url, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json;charset=utf-8'
                },
                body: JSON.stringify(del.toJSON()),
                ...options,
            });
            const buffer = await response.arrayBuffer();
            return planPB.Plan.decode(new Uint8Array(buffer));
        },
    }
}

export const APIContext = createContext<API | null>(null);
