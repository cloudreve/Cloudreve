import {
    combineReducers as combine,
    ReducersMapObject,
    AnyAction,
} from "redux";
import invariant from "invariant";

export const combineReducers = (reducers: ReducersMapObject) => {
    const combinedReducer = combine(reducers);
    // TODO: define state type
    return (state: any, action: AnyAction) => {
        if (
            action.type &&
            !action.type.startsWith("@@") &&
            action.type.split("/").length > 1
        ) {
            const namespace = action.type.split("/")[0];
            const reducer = reducers[namespace];
            invariant(!!reducer, `reducer ${namespace} doesn't exist`);
            return reducer && reducer(state, action);
        }
        return combinedReducer(state, action);
    };
};
