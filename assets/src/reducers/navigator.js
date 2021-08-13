const navigator = (state = [], action) => {
    switch (action.type) {
        case "NAVIGATOR_TO":
            return Object.assign({}, state, {
                path: action.path,
            });
        default:
            return state;
    }
};

export default navigator;
