import { getProfile } from "@/api/profile_service";

export function Profile() {
    return (<>{
        getProfile("me").then(x => {
            console.log(x)
            return (<>
                <p>{x.id}</p>
                <p>{x.first_name}</p>
                <p>{x.last_name}</p>
                <p>{x.role}</p>
                <p>{x.picture_url}</p>
                <p>{x.user_status}</p>
            </>)
        }).catch(e => {
            return <p>{JSON.stringify(e)}</p>
        })
    }
    </>
    )
}
